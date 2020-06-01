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

type ProducerMessage struct {
	Rooms   []int32
	User    User
	Display Display
	Type    string
	IsSave  bool
}

func (p *Producer) toPb(msg ProducerMessage) (*logicpb.PushMsg, error) {
	if err := checkMessage(msg.Display.Message.Text); err != nil {
		return nil, err
	}

	seq, err := p.seq.Id(context.Background(), &seqpb.SeqReq{
		Id: 1, Count: 1,
	})
	if err != nil {
		return nil, err
	}

	fmsg, isMatch, sensitive := p.filter.FilterFindSensitive(msg.Display.Message.Text)
	if isMatch {
		log.Info("message filter hit", zap.Int64("msg_id", seq.Id), zap.Strings("sensitive", sensitive))
	}

	msg.Display.Message.Text = fmsg

	now := time.Now()
	bm, err := json.Marshal(Message{
		Id:        seq.Id,
		Type:      msg.Type,
		Time:      now.Format("15:04:05"),
		Timestamp: now.Unix(),
		User:      NullUser(msg.User),
		Display:   msg.Display,

		Uid:     msg.User.Uid,
		Name:    msg.Display.User.Text,
		Avatar:  msg.Display.User.Avatar,
		Message: msg.Display.Message.Text,
	})
	if err != nil {
		return nil, err
	}
	return &logicpb.PushMsg{
		Seq:     seq.Id,
		Type:    logicpb.PushMsg_USER,
		Room:    msg.Rooms,
		Mid:     msg.User.Id,
		Msg:     bm,
		Message: msg.Display.Message.Text,
		SendAt:  now.Unix(),
		IsSave:  msg.IsSave,
	}, nil
}

func (p *Producer) Send(msg ProducerMessage) (int64, error) {
	if err := p.rate.perSec(msg.User.Id); err != nil {
		return 0, err
	}
	if err := p.rate.sameMsg(msg); err != nil {
		return 0, err
	}

	msg.Type = MessageType
	pushMsg, err := p.toPb(msg)
	if err != nil {
		return 0, err
	}

	pushMsg.IsRaw = true

	if err := p.send(pushMsg); err != nil {
		return 0, err
	}
	return pushMsg.Seq, nil
}

func (p *Producer) SendForAdmin(msg ProducerMessage) (int64, error) {
	msg.Type = MessageType
	pushMsg, err := p.toPb(msg)

	if err != nil {
		return 0, err
	}

	pushMsg.Type = logicpb.PushMsg_ADMIN

	if err := p.send(pushMsg); err != nil {
		return 0, err
	}
	return pushMsg.Seq, nil
}

func (p *Producer) SendTop(msg ProducerMessage) (int64, error) {
	msg.Type = TopType
	pushMsg, err := p.toPb(msg)

	if err != nil {
		return 0, err
	}

	pushMsg.Type = logicpb.PushMsg_ADMIN_TOP

	if err := p.send(pushMsg); err != nil {
		return 0, err
	}
	return pushMsg.Seq, nil
}

func (p *Producer) SendGift(rid int32, user User, gift Gift) (int64, error) {
	seq, err := p.seq.Id(context.Background(), &seqpb.SeqReq{
		Id: 1, Count: 1,
	})
	if err != nil {
		return 0, err
	}

	gift.Message = "送出" + gift.Name

	now := time.Now()
	bm, err := json.Marshal(GiftMessage{
		Message: Message{
			Id:        seq.Id,
			Type:      GiftType,
			Time:      now.Format("15:04:05"),
			Timestamp: now.Unix(),
			Display:   DisplayByGift(user, gift.Name),
			User:      NullUser(user),
		},
		Gift: gift,
	})

	if err != nil {
		return 0, err
	}

	pushMsg := &logicpb.PushMsg{
		Seq:    seq.Id,
		Type:   logicpb.PushMsg_SYSTEM,
		Room:   []int32{rid},
		Mid:    user.Id,
		Msg:    bm,
		SendAt: now.Unix(),
		IsRaw:  true,
	}
	if err := p.send(pushMsg); err != nil {
		return 0, err
	}
	return pushMsg.Seq, nil
}

func (p *Producer) toRedEnvelopePb(msg ProducerMessage, redEnvelope RedEnvelope) (*logicpb.PushMsg, error) {
	if err := checkMessage(msg.Display.Message.Text); err != nil {
		return nil, err
	}
	seq, err := p.seq.Id(context.Background(), &seqpb.SeqReq{
		Id: 1, Count: 1,
	})
	if err != nil {
		return nil, err
	}

	fmsg, isMatch, sensitive := p.filter.FilterFindSensitive(msg.Display.Message.Text)
	if isMatch {
		log.Info("message filter hit", zap.Int64("msg_id", seq.Id), zap.Strings("sensitive", sensitive))
	}

	msg.Display.Message.Text = fmsg

	now := time.Now()
	bm, err := json.Marshal(RedEnvelopeMessage{
		Message: Message{
			Id:        seq.Id,
			Type:      RedEnvelopeType,
			Time:      now.Format("15:04:05"),
			Timestamp: now.Unix(),
			User:      NullUser(msg.User),
			Display:   msg.Display,

			Uid:     msg.User.Uid,
			Name:    msg.Display.User.Text,
			Avatar:  msg.Display.User.Avatar,
			Message: msg.Display.Message.Text,
		},
		RedEnvelope: redEnvelope,
	})
	if err != nil {
		return nil, err
	}
	return &logicpb.PushMsg{
		Seq:     seq.Id,
		Type:    logicpb.PushMsg_MONEY,
		Room:    msg.Rooms,
		Mid:     msg.User.Id,
		Msg:     bm,
		SendAt:  now.Unix(),
		Message: msg.Display.Message.Text,
		IsSave:  msg.IsSave,
	}, nil
}

func (p *Producer) SendRedEnvelope(msg ProducerMessage, redEnvelope RedEnvelope) (int64, error) {
	pushMsg, err := p.toRedEnvelopePb(msg, redEnvelope)
	if err != nil {
		return 0, err
	}
	if err := p.send(pushMsg); err != nil {
		return 0, err
	}
	return pushMsg.Seq, nil
}

func (p *Producer) SendBets(msg ProducerMessage, bet Bet) (int64, error) {
	seq, err := p.seq.Id(context.Background(), &seqpb.SeqReq{
		Id: 1, Count: 1,
	})
	if err != nil {
		return 0, err
	}

	// 避免Items與TransItems欄位json Marshal後出現null
	for i, v := range bet.Orders {
		if len(v.Items) == 0 {
			bet.Orders[i].Items = []string{}
		}
		if len(v.TransItems) == 0 {
			bet.Orders[i].TransItems = []string{}
		}
	}

	now := time.Now()
	bm, err := json.Marshal(Bets{
		Id:        seq.Id,
		Type:      BetsType,
		Time:      now.Format("15:04:05"),
		Timestamp: now.Unix(),
		Display:   msg.Display,
		User:      msg.User,
		Bet:       bet,

		Uid:          msg.User.Uid,
		Name:         msg.User.Name,
		Avatar:       msg.User.Avatar,
		GameId:       bet.GameId,
		PeriodNumber: bet.PeriodNumber,
		Items:        bet.Orders,
		Count:        bet.Count,
		TotalAmount:  bet.TotalAmount,
	})
	if err != nil {
		return 0, err
	}

	pushMsg := &logicpb.PushMsg{
		Seq:    seq.Id,
		Type:   logicpb.PushMsg_SYSTEM,
		Room:   msg.Rooms,
		Mid:    msg.User.Id,
		Msg:    bm,
		SendAt: now.Unix(),
	}
	if err := p.send(pushMsg); err != nil {
		return 0, err
	}
	return pushMsg.Seq, nil
}

func (p *Producer) SendRaw(roomId []int32, body []byte, IsRaw bool) (int64, error) {
	seq, err := p.seq.Id(context.Background(), &seqpb.SeqReq{
		Id: 1, Count: 1,
	})
	if err != nil {
		return 0, err
	}

	now := time.Now()
	var b map[string]interface{}
	if err := json.Unmarshal(body, &b); err != nil {
		return 0, err
	}

	b["id"] = seq.Id
	b["time"] = now.Format("15:04:05")
	b["timestamp"] = now.Unix()

	bm, err := json.Marshal(b)
	if err != nil {
		return 0, err
	}

	pushMsg := &logicpb.PushMsg{
		Seq:    seq.Id,
		Type:   logicpb.PushMsg_SYSTEM,
		Room:   roomId,
		Msg:    bm,
		SendAt: now.Unix(),
		IsRaw:  IsRaw,
	}
	if err := p.send(pushMsg); err != nil {
		return 0, err
	}
	return pushMsg.Seq, nil
}

type RawMessage struct {
	// 訊息房間
	RoomId []int32 `json:"room_id" binding:"required"`

	// 訊息資料
	Body string `json:"body" binding:"required"`
}

func (p *Producer) SendRaws(raws []RawMessage, IsRaw bool) (int64, error) {
	count := int64(len(raws))
	seq, err := p.seq.Ids(context.Background(), &seqpb.SeqReq{
		Id: 1, Count: count,
	})
	if err != nil {
		return 0, err
	}

	now := time.Now()
	id := seq.Id - count

	for i, raw := range raws {
		var b map[string]interface{}
		if err := json.Unmarshal([]byte(raw.Body), &b); err != nil {
			return 0, err
		}

		seq := id + int64(i) + 1
		b["id"] = seq
		b["time"] = now.Format("15:04:05")
		b["timestamp"] = now.Unix()

		bm, err := json.Marshal(b)
		if err != nil {
			return 0, err
		}

		pushMsg := &logicpb.PushMsg{
			Seq:    seq,
			Type:   logicpb.PushMsg_SYSTEM,
			Room:   raw.RoomId,
			Msg:    bm,
			SendAt: now.Unix(),
			IsRaw:  IsRaw,
		}

		if err := p.send(pushMsg); err != nil {
			if i == 0 {
				return 0, err
			}
			return id, err
		}
	}

	return id, nil
}

func (p *Producer) Kick(msg string, keys []string) error {
	pushMsg := &logicpb.PushMsg{
		Type:    logicpb.PushMsg_Close,
		Keys:    keys,
		Message: msg,
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
		log.Error(
			"message producer send message",
			zap.Error(err),
			zap.Int32("partition", partition),
			zap.Int64("offset", offset),
			zap.String("topic", p.topic),
			zap.String("type", pushMsg.Type.String()),
			zap.String("msg", pushMsg.Message),
			zap.Int32s("rooms", pushMsg.Room),
		)
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
