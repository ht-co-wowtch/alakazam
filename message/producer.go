package message

import (
	"context"
	"encoding/json"
	kafka "github.com/Shopify/sarama"
	"github.com/go-redis/redis"
	"github.com/gogo/protobuf/proto"
	logicpb "gitlab.com/jetfueltw/cpw/alakazam/app/logic/pb"
	seqpb "gitlab.com/jetfueltw/cpw/alakazam/app/seq/api/pb"
	"gitlab.com/jetfueltw/cpw/alakazam/models"
	shield "gitlab.com/jetfueltw/cpw/alakazam/pkg/filter"
	"gitlab.com/jetfueltw/cpw/micro/log"
	"go.uber.org/zap"
	"time"
)

type Producer struct {
	topic    string
	brokers  []string
	producer kafka.SyncProducer
	seq      seqpb.SeqClient
	rate     *rateLimit
	filter   *shield.Filter
	bs       map[int64]*seq
}

type seq struct {
	cur int64
	max int64
}

func NewProducer(brokers []string, topic string, seq seqpb.SeqClient, cache *redis.Client, db *models.Store) *Producer {
	kc := kafka.NewConfig()
	kc.Version = kafka.V2_3_0_0
	kc.Producer.Return.Successes = true
	pub, err := kafka.NewSyncProducer(brokers, kc)
	if err != nil {
		panic(err)
	}

	f := &filter{
		db:     db,
		filter: shield.New(),
	}
	if err := f.load(); err != nil {
		panic(err)
	}

	go func() {
		t := time.NewTicker(time.Hour)
		for {
			select {
			case <-t.C:
				if err := f.load(); err != nil {
					log.Error("reload shield", zap.Error(err))
				}
			}
		}
	}()

	return &Producer{
		brokers:  brokers,
		topic:    topic,
		producer: pub,
		seq:      seq,
		filter:   f.filter,
		rate:     newRateLimit(cache),
	}
}

func (p *Producer) Close() error {
	return p.producer.Close()
}

const (
	RootMid  = 1
	RootUid  = "root"
	RootName = "管理员"
)

type Messages struct {
	Rooms   []int32
	Mid     int64
	Uid     string
	Name    string
	Message string
	IsTop   bool
	Type    string
}

func (p *Producer) toPb(msg Messages) (*logicpb.PushMsg, error) {
	seq, err := p.seq.Id(context.Background(), &seqpb.SeqReq{
		Id: 1, Count: 1,
	})
	if err != nil {
		return nil, err
	}

	fmsg, isMatch, sensitive := p.filter.FilterFindSensitive(msg.Message)
	if isMatch {
		log.Info("message filter hit", zap.Int64("msg_id", seq.Id), zap.Strings("sensitive", sensitive))
	}

	now := time.Now()
	bm, err := json.Marshal(Message{
		Id:      seq.Id,
		Type:    msg.Type,
		Uid:     msg.Uid,
		Name:    msg.Name,
		Message: fmsg,
		Time:    now.Format("15:04:05"),
	})
	if err != nil {
		return nil, err
	}
	return &logicpb.PushMsg{
		Seq:     seq.Id,
		Type:    logicpb.PushMsg_ROOM,
		Room:    msg.Rooms,
		Mid:     msg.Mid,
		Msg:     bm,
		Message: fmsg,
		SendAt:  now.Unix(),
	}, nil
}

func (p *Producer) Send(msg Messages) (int64, error) {
	if err := p.rate.perSec(msg.Mid); err != nil {
		return 0, err
	}
	if err := p.rate.sameMsg(msg); err != nil {
		return 0, err
	}

	msg.Type = messageType
	pushMsg, err := p.toPb(msg)
	if err != nil {
		return 0, err
	}
	if err := p.send(pushMsg); err != nil {
		return 0, err
	}
	return pushMsg.Seq, nil
}

type AdminMessage struct {
	Rooms   []int32
	Message string
	IsTop   bool
}

// 所有房間推送
// TODO 需實作訊息是否頂置
func (p *Producer) SendForAdmin(msg AdminMessage) (int64, error) {
	ty := messageType
	if msg.IsTop {
		ty = topType
	}
	pushMsg, err := p.toPb(Messages{
		Rooms:   msg.Rooms,
		Mid:     RootMid,
		Uid:     RootUid,
		Name:    RootName,
		Message: msg.Message,
		Type:    ty,
	})
	if err != nil {
		return 0, err
	}
	if msg.IsTop {
		pushMsg.Type = logicpb.PushMsg_TOP
	}
	if err := p.send(pushMsg); err != nil {
		return 0, err
	}
	return pushMsg.Seq, nil
}

type KickMessage struct {
	Message string
	Keys    []string
}

func (p *Producer) Kick(msg KickMessage) error {
	pushMsg := &logicpb.PushMsg{
		Type:    logicpb.PushMsg_Close,
		Keys:    msg.Keys,
		Message: msg.Message,
	}
	if err := p.send(pushMsg); err != nil {
		return err
	}
	return nil
}

type RedEnvelopeMessage struct {
	Messages
	RedEnvelopeId string
	Token         string
	Expired       time.Time
}

func (p *Producer) toRedEnvelopePb(msg RedEnvelopeMessage) (*logicpb.PushMsg, error) {
	seq, err := p.seq.Id(context.Background(), &seqpb.SeqReq{
		Id: 1, Count: 1,
	})
	if err != nil {
		return nil, err
	}

	fmsg, isMatch, sensitive := p.filter.FilterFindSensitive(msg.Message)
	if isMatch {
		log.Info("message filter hit", zap.Int64("msg_id", seq.Id), zap.Strings("sensitive", sensitive))
	}

	now := time.Now()
	bm, err := json.Marshal(Money{
		Message: Message{
			Id:      seq.Id,
			Type:    redEnvelopeType,
			Uid:     msg.Uid,
			Name:    msg.Name,
			Message: fmsg,
			Time:    now.Format("15:04:05"),
		},
		RedEnvelope: RedEnvelope{
			Id:      msg.RedEnvelopeId,
			Token:   msg.Token,
			Expired: msg.Expired.Format(time.RFC3339),
		},
	})
	if err != nil {
		return nil, err
	}
	return &logicpb.PushMsg{
		Seq:     seq.Id,
		Type:    logicpb.PushMsg_MONEY,
		Room:    msg.Rooms,
		Mid:     msg.Mid,
		Msg:     bm,
		SendAt:  now.Unix(),
		Message: fmsg,
	}, nil
}

func (p *Producer) SendRedEnvelope(msg RedEnvelopeMessage) (int64, error) {
	pushMsg, err := p.toRedEnvelopePb(msg)
	if err != nil {
		return 0, err
	}
	if err := p.send(pushMsg); err != nil {
		return 0, err
	}
	return pushMsg.Seq, nil
}

type AdminRedEnvelopeMessage struct {
	AdminMessage
	RedEnvelopeId string
	Token         string
	Expired       time.Time
}

func (p *Producer) SendRedEnvelopeForAdmin(msg AdminRedEnvelopeMessage) (int64, error) {
	pushMsg, err := p.toRedEnvelopePb(RedEnvelopeMessage{
		Messages: Messages{
			Rooms:   msg.Rooms,
			Mid:     RootMid,
			Uid:     RootUid,
			Name:    RootName,
			Message: msg.Message,
		},
		RedEnvelopeId: msg.RedEnvelopeId,
		Token:         msg.Token,
		Expired:       msg.Expired,
	})
	if err != nil {
		return 0, err
	}
	if err := p.send(pushMsg); err != nil {
		return 0, err
	}
	return pushMsg.Seq, nil
}

// 房間推送，以下為條件
// 1. room id
func (p *Producer) send(pushMsg *logicpb.PushMsg) error {
	b, err := proto.Marshal(pushMsg)
	if err != nil {
		return err
	}
	var key kafka.StringEncoder
	switch pushMsg.Type {
	case logicpb.PushMsg_ROOM:
		key = kafka.StringEncoder(pushMsg.Room[0])
	case logicpb.PushMsg_MONEY:
		key = redEnvelopeType
	case logicpb.PushMsg_TOP:
		key = topType
	}
	m := &kafka.ProducerMessage{
		Key:   key,
		Topic: p.topic,
		Value: kafka.ByteEncoder(b),
	}
	partition, offset, err := p.producer.SendMessage(m)
	if err != nil {
		log.Error("message producer send message", zap.Error(err), zap.Int32("partition", partition), zap.Int64("offset", offset))
	}
	return err
}
