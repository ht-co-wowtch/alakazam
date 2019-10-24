package message

import (
	"context"
	"encoding/json"
	"fmt"
	kafka "github.com/Shopify/sarama"
	"github.com/go-redis/redis"
	"github.com/gogo/protobuf/proto"
	logicpb "gitlab.com/jetfueltw/cpw/alakazam/app/logic/pb"
	seqpb "gitlab.com/jetfueltw/cpw/alakazam/app/seq/api/pb"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"gitlab.com/jetfueltw/cpw/alakazam/models"
	shield "gitlab.com/jetfueltw/cpw/alakazam/pkg/filter"
	"gitlab.com/jetfueltw/cpw/micro/log"
	"go.uber.org/zap"
	"golang.org/x/net/html"
	"regexp"
	"strings"
	"time"
)

type Producer struct {
	topic    string
	brokers  []string
	producer kafka.SyncProducer
	cache    *cache
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
	// 參數對照https://kafka.apache.org/documentation/#producerconfigs
	kc.Producer.Return.Successes = true
	// acks
	kc.Producer.RequiredAcks = kafka.WaitForLocal
	// compression.type
	kc.Producer.Compression = kafka.CompressionNone
	// retries
	kc.Producer.Retry.Max = 1
	// max.in.flight.requests.per.connection
	kc.Net.MaxOpenRequests = 1
	// max.request.size，需小於或等於 broker `message.max.bytes`
	kc.Producer.MaxMessageBytes = 1000000
	// Producer 等待多少Bytes後再一併發送給broker
	kc.Producer.Flush.Bytes = 0
	// linger.ms
	kc.Producer.Flush.Frequency = time.Duration(0)
	// batch.size
	kc.Producer.Flush.Messages = 0
	// Producer 最大緩衝訊息筆數
	kc.Producer.Flush.MaxMessages = 0
	// request.timeout.ms
	kc.Producer.Timeout = 10 * time.Second

	client, err := kafka.NewClient(brokers, kc)
	if err != nil {
		panic(err)
	}

	pub, err := kafka.NewSyncProducerFromClient(client)
	if err != nil {
		panic(err)
	}

	if err := registerProducerMetric(client, kc.MetricRegistry); err != nil {
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
		cache:    newCache(cache),
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

type ProducerMessage struct {
	Rooms   []int32
	Mid     int64
	Uid     string
	Name    string
	Message string
	IsTop   bool
	Type    string
	Avatar  int
}

func (p *Producer) toPb(msg ProducerMessage) (*logicpb.PushMsg, error) {
	if err := checkMessage(msg.Message); err != nil {
		return nil, err
	}

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
		Id:        seq.Id,
		Type:      msg.Type,
		Uid:       msg.Uid,
		Name:      msg.Name,
		Avatar:    toAvatarName(msg.Avatar),
		Message:   fmsg,
		Time:      now.Format("15:04:05"),
		Timestamp: now.Unix(),
	})
	if err != nil {
		return nil, err
	}
	return &logicpb.PushMsg{
		Seq:     seq.Id,
		Type:    logicpb.PushMsg_USER,
		Room:    msg.Rooms,
		Mid:     msg.Mid,
		Msg:     bm,
		Message: fmsg,
		SendAt:  now.Unix(),
	}, nil
}

func (p *Producer) Send(msg ProducerMessage) (int64, error) {
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

type ProducerAdminMessage struct {
	Rooms   []int32
	Name    string
	Message string
	IsTop   bool
}

// 所有房間推送
func (p *Producer) SendForAdmin(msg ProducerAdminMessage) (int64, error) {
	ty := messageType
	if msg.IsTop {
		ty = TopType
	}
	pushMsg, err := p.toPb(ProducerMessage{
		Rooms:   msg.Rooms,
		Mid:     RootMid,
		Uid:     RootUid,
		Name:    RootName,
		Avatar:  99,
		Message: msg.Message,
		Type:    ty,
	})
	if err != nil {
		return 0, err
	}
	if msg.IsTop {
		pushMsg.Type = logicpb.PushMsg_ADMIN_TOP
	} else {
		pushMsg.Type = logicpb.PushMsg_ADMIN
	}
	if err := p.send(pushMsg); err != nil {
		return 0, err
	}
	return pushMsg.Seq, nil
}

type ProducerKickMessage struct {
	Message string
	Keys    []string
}

func (p *Producer) Kick(msg ProducerKickMessage) error {
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

func (p *Producer) CloseTop(msgId int64, rid []int32) error {
	pushMsg := &logicpb.PushMsg{
		Type: logicpb.PushMsg_CLOSE_TOP,
		Seq:  msgId,
		Room: rid,
		Msg:  []byte(fmt.Sprintf(`{"id":%d}`, msgId)),
	}
	if err := p.send(pushMsg); err != nil {
		return err
	}
	return nil
}

type ProducerRedEnvelopeMessage struct {
	ProducerMessage
	RedEnvelopeId string
	Token         string
	Expired       time.Time
}

func (p *Producer) toRedEnvelopePb(msg ProducerRedEnvelopeMessage) (*logicpb.PushMsg, error) {
	if err := checkMessage(msg.Message); err != nil {
		return nil, err
	}
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
	bm, err := json.Marshal(RedEnvelopeMessage{
		Message: Message{
			Id:        seq.Id,
			Type:      redEnvelopeType,
			Uid:       msg.Uid,
			Name:      msg.Name,
			Avatar:    toAvatarName(msg.Avatar),
			Message:   fmsg,
			Time:      now.Format("15:04:05"),
			Timestamp: now.Unix(),
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

func (p *Producer) SendRedEnvelope(msg ProducerRedEnvelopeMessage) (int64, error) {
	pushMsg, err := p.toRedEnvelopePb(msg)
	if err != nil {
		return 0, err
	}
	if err := p.send(pushMsg); err != nil {
		return 0, err
	}
	return pushMsg.Seq, nil
}

type ProducerAdminRedEnvelopeMessage struct {
	ProducerAdminMessage
	RedEnvelopeId string
	Token         string
	Expired       time.Time
}

func (p *Producer) SendRedEnvelopeForAdmin(msg ProducerAdminRedEnvelopeMessage) (int64, error) {
	pushMsg, err := p.toRedEnvelopePb(ProducerRedEnvelopeMessage{
		ProducerMessage: ProducerMessage{
			Rooms:   msg.Rooms,
			Mid:     RootMid,
			Uid:     RootUid,
			Name:    msg.Name,
			Message: msg.Message,
			Avatar:  99,
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

type ProducerBetsMessage struct {
	Rooms  []int32
	Mid    int64
	Uid    string
	Name   string
	Avatar int

	GameId       int
	PeriodNumber int
	Bets         []Bet
	Count        int
	TotalAmount  int
}

func (p *Producer) SendBets(msg ProducerBetsMessage) (int64, error) {
	seq, err := p.seq.Id(context.Background(), &seqpb.SeqReq{
		Id: 1, Count: 1,
	})
	if err != nil {
		return 0, err
	}

	now := time.Now()
	bm, err := json.Marshal(Bets{
		Id:           seq.Id,
		Type:         betsType,
		Uid:          msg.Uid,
		Name:         msg.Name,
		Avatar:       toAvatarName(msg.Avatar),
		Time:         now.Format("15:04:05"),
		Timestamp:    now.Unix(),
		GameId:       msg.GameId,
		PeriodNumber: msg.PeriodNumber,
		Items:        msg.Bets,
		Count:        msg.Count,
		TotalAmount:  msg.TotalAmount,
	})
	if err != nil {
		return 0, err
	}

	pushMsg := &logicpb.PushMsg{
		Seq:    seq.Id,
		Type:   logicpb.PushMsg_BETS,
		Room:   msg.Rooms,
		Mid:    msg.Mid,
		Msg:    bm,
		SendAt: now.Unix(),
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
	case logicpb.PushMsg_Close:
		key = kafka.StringEncoder(pushMsg.Keys[0])
	default:
		key = kafka.StringEncoder(pushMsg.Room[0])
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

var (
	// 空白字元
	msgRegex = regexp.MustCompile(`^(\s|\xE3\x80\x80)*$`)
)

func checkMessage(msg string) error {
	if msgRegex.MatchString(msg) {
		return errors.ErrIllegal
	}

	var textCount uint8
	tokenizer := html.NewTokenizer(strings.NewReader(msg))
	for {
		if tokenizer.Next() == html.ErrorToken {
			if textCount == 0 {
				return errors.ErrIllegal
			}
			return nil
		}

		token := tokenizer.Token()
		switch token.Type {
		case html.StartTagToken, html.EndTagToken, html.SelfClosingTagToken:
			return errors.ErrIllegal
		case html.TextToken:
			textCount++
			break
		default:
			break
		}
	}
}
