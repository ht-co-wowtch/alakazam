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
	"gitlab.com/jetfueltw/cpw/alakazam/message/scheme"
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
	User    scheme.User
	Display scheme.Display
	Type    string
	IsSave  bool
}

func (p *Producer) SendUser(rid []int32, msg string, user *models.Member) (int64, error) {
	if err := p.rate.perSec(user.Id); err != nil {
		return 0, err
	}
	if err := p.rate.sameMsg(msg, user.Uid); err != nil {
		return 0, err
	}

	msg, err := p.filterMessage(msg)

	var id int64
	if id, err = p.id(); err != nil {
		return 0, err
	}

	u := scheme.User{
		Id:     user.Id,
		Uid:    user.Uid,
		Name:   user.Name,
		Avatar: ToAvatarName(user.Gender),
	}

	var message scheme.Message
	if user.Type == models.STREAMER {
		message = u.ToStreamer(id, msg)
	} else {
		message = u.ToUser(id, msg)
	}

	pushMsg, err := message.ToPb(user.Id, rid, logicpb.PushMsg_USER, true, true)
	if err != nil {
		return 0, err
	}

	if err := p.send(pushMsg); err != nil {
		return 0, err
	}
	return pushMsg.Seq, nil
}

func (p *Producer) SendSystem(rid []int32, msg string) (int64, error) {
	id, err := p.id()
	if err != nil {
		return 0, err
	}

	u := scheme.NewRoot()

	pushMsg, err := u.ToSystem(id, msg).ToPb(u.Id, rid, logicpb.PushMsg_USER, false, false)
	if err != nil {
		return 0, err
	}

	if err := p.send(pushMsg); err != nil {
		return 0, err
	}
	return pushMsg.Seq, nil
}

func (p *Producer) SendAdmin(rid []int32, msg string) (int64, error) {
	id, err := p.id()
	if err != nil {
		return 0, err
	}

	u := scheme.NewRoot()

	pushMsg, err := u.ToAdmin(id, msg).ToPb(u.Id, rid, logicpb.PushMsg_ADMIN, false, false)
	if err != nil {
		return 0, err
	}

	if err := p.send(pushMsg); err != nil {
		return 0, err
	}
	return pushMsg.Seq, nil
}

func (p *Producer) SendTop(rid []int32, msg string) (int64, error) {
	id, err := p.id()
	if err != nil {
		return 0, err
	}

	u := scheme.NewRoot()

	pushMsg, err := u.ToTop(id, msg).ToPb(u.Id, rid, logicpb.PushMsg_ADMIN_TOP, false, false)
	if err != nil {
		return 0, err
	}

	if err := p.send(pushMsg); err != nil {
		return 0, err
	}
	return pushMsg.Seq, nil
}

func (p *Producer) SendGift(rid int32, user scheme.User, gift scheme.Gift) (int64, error) {
	id, err := p.id()
	if err != nil {
		return 0, err
	}

	if gift.Combo.Count == 0 {
		gift.ShowAnimation = true
	}

	pushMsg, err := gift.ToMessage(id, user).ToPb(user.Id, rid)
	if err != nil {
		return 0, err
	}

	if err := p.send(pushMsg); err != nil {
		return 0, err
	}
	return pushMsg.Seq, nil
}

func (p *Producer) SendReward(rid int32, user scheme.User, amount, totalAmount float64) (int64, error) {
	id, err := p.id()
	if err != nil {
		return 0, err
	}

	pushMsg, err := scheme.NewReward(id, user, amount, totalAmount).ToPb(user.Id, rid)
	if err != nil {
		return 0, err
	}

	if err := p.send(pushMsg); err != nil {
		return 0, err
	}
	return pushMsg.Seq, nil
}

func (p *Producer) SendRedEnvelope(rid []int32, message string, user scheme.User, redEnvelope scheme.RedEnvelope) (int64, error) {
	pushMsg, err := p.toRedEnvelopePb(rid, message, user, redEnvelope)
	if err != nil {
		return 0, err
	}
	if err := p.send(pushMsg); err != nil {
		return 0, err
	}
	return pushMsg.Seq, nil
}

func (p *Producer) SendBets(rid []int32, user scheme.User, bet scheme.Bet) (int64, error) {
	id, err := p.id()
	if err != nil {
		return 0, err
	}

	message := bet.ToMessage(id, user)

	bm, err := json.Marshal(message)
	if err != nil {
		return 0, err
	}

	pushMsg := &logicpb.PushMsg{
		Seq:     id,
		Type:    logicpb.PushMsg_SYSTEM,
		Room:    rid,
		Mid:     user.Id,
		Msg:     bm,
		Message: message.Display.Message.Text,
		SendAt:  message.Timestamp,
		IsSave:  false,
		IsRaw:   false,
	}

	if err := p.send(pushMsg); err != nil {
		return 0, err
	}
	return pushMsg.Seq, nil
}

func (p *Producer) SendBetsWin(rid []int32, user scheme.User, gameName string) (int64, error) {
	id, err := p.id()
	if err != nil {
		return 0, err
	}

	pushMsg, err := scheme.NewBetsWin(id, user, gameName).ToPb(user.Id, rid, logicpb.PushMsg_SYSTEM, false, false)
	if err != nil {
		return 0, err
	}

	if err := p.send(pushMsg); err != nil {
		return 0, err
	}
	return pushMsg.Seq, nil
}

func (p *Producer) SendBetsWinReward(keys []string, user scheme.User, amount float64, buttonName string) (int64, error) {
	id, err := p.id()
	if err != nil {
		return 0, err
	}

	pushMsg, err := scheme.NewBetsWinReward(id, user, amount, buttonName).ToPb(keys)
	if err != nil {
		return 0, err
	}

	if err := p.send(pushMsg); err != nil {
		return 0, err
	}
	return pushMsg.Seq, nil
}

func (p *Producer) SendRaw(roomId []int32, body []byte, IsRaw bool) (int64, error) {
	id, err := p.id()
	if err != nil {
		return 0, err
	}

	now := time.Now()
	var b map[string]interface{}
	if err := json.Unmarshal(body, &b); err != nil {
		return 0, err
	}

	b["id"] = id
	b["time"] = now.Format("15:04:05")
	b["timestamp"] = now.Unix()

	bm, err := json.Marshal(b)
	if err != nil {
		return 0, err
	}

	pushMsg := &logicpb.PushMsg{
		Seq:    id,
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
	id, err := p.id()
	if err != nil {
		return 0, err
	}

	now := time.Now()
	id = id - count

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

func (p *Producer) SendConnect(rid int32, user *logicpb.User) (int64, error) {
	id, err := p.id()
	if err != nil {
		return 0, err
	}

	pushMsg, err := scheme.NewConnect(id, user.Name).ToPb(user.Id, []int32{rid}, logicpb.PushMsg_SYSTEM, false, false)
	if err != nil {
		return 0, err
	}

	if err := p.send(pushMsg); err != nil {
		return 0, err
	}
	return pushMsg.Seq, nil
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

func (p *Producer) id() (int64, error) {
	seq, err := p.seq.Id(context.Background(), &seqpb.SeqReq{
		Id: 1, Count: 1,
	})
	if err != nil {
		return 0, err
	}
	return seq.Id, err
}

func (p *Producer) filterMessage(message string) (string, error) {
	if err := checkMessage(message); err != nil {
		return "", err
	}

	fmsg, isMatch, sensitive := p.filter.FilterFindSensitive(message)
	if isMatch {
		log.Info("message filter hit", zap.Strings("sensitive", sensitive))
	}

	return fmsg, nil
}

func (p *Producer) toRedEnvelopePb(rid []int32, message string, user scheme.User, redEnvelope scheme.RedEnvelope) (*logicpb.PushMsg, error) {
	message, err := p.filterMessage(message)
	if err != nil {
		return nil, err
	}

	id, err := p.id()
	if err != nil {
		return nil, err
	}

	msg := redEnvelope.ToMessage(id, message, user)

	bm, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}

	return &logicpb.PushMsg{
		Seq:     id,
		Type:    logicpb.PushMsg_MONEY,
		Room:    rid,
		Mid:     user.Id,
		Msg:     bm,
		Message: message,
		SendAt:  msg.Timestamp,
		IsSave:  true,
		IsRaw:   false,
	}, nil
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
