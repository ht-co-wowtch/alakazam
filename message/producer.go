package message

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	kafka "github.com/Shopify/sarama"
	"github.com/go-redis/redis"
	"github.com/gogo/protobuf/proto"
	"gitlab.com/ht-co/cpw/micro/log"
	"gitlab.com/ht-co/wowtch/live/alakazam/app/comet/pb"
	logicpb "gitlab.com/ht-co/wowtch/live/alakazam/app/logic/pb"
	seqpb "gitlab.com/ht-co/wowtch/live/alakazam/app/seq/api/pb"
	"gitlab.com/ht-co/wowtch/live/alakazam/errors"
	"gitlab.com/ht-co/wowtch/live/alakazam/message/scheme"
	"gitlab.com/ht-co/wowtch/live/alakazam/models"
	shield "gitlab.com/ht-co/wowtch/live/alakazam/pkg/filter"
	"go.uber.org/zap"
	"golang.org/x/net/html"
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

func NewKafkaProducer(brokers []string) kafka.SyncProducer {
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

	return pub
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

func (p *Producer) Send(fun func(id int64) (*logicpb.PushMsg, error)) (int64, error) {
	id, err := p.id()
	if err != nil {
		return 0, err
	}

	pushMsg, err := fun(id)
	if err != nil {
		return 0, err
	}

	if err := p.send(pushMsg); err != nil {
		return 0, err
	}
	return pushMsg.Seq, nil
}

func (p *Producer) SendKey(keys []string, msg string, user *models.Member) (int64, error) {
	if err := p.rate.perSec(user.Id); err != nil {
		return 0, err
	}
	if err := p.rate.sameMsg(msg, user.Uid); err != nil {
		return 0, err
	}

	msg, err := p.filterMessage(msg)
	if err != nil {
		return 0, err
	}

	return p.Send(func(id int64) (*logicpb.PushMsg, error) {
		pushMsg, err := scheme.NewUser(*user).ToUser(id, msg).ToProto()
		if err != nil {
			return nil, err
		}

		pushMsg.Keys = keys
		pushMsg.Mid = user.Id
		pushMsg.Message = msg
		pushMsg.Type = logicpb.PushMsg_PUSH
		pushMsg.MsgType = models.MESSAGE_TYPE
		pushMsg.IsRaw = true

		return pushMsg, nil
	})
}

// 發送訊息
func (p *Producer) SendUser(rid []int32, msg string, user *models.Member) (int64, error) {
	if err := p.rate.perSec(user.Id); err != nil {
		return 0, err
	}
	if err := p.rate.sameMsg(msg, user.Uid); err != nil {
		return 0, err
	}

	msg, err := p.filterMessage(msg)
	if err != nil {
		return 0, err
	}

	return p.Send(func(id int64) (*logicpb.PushMsg, error) {
		u := scheme.NewUser(*user)

		var message scheme.Message

		if user.Type == models.STREAMER {
			message = u.ToStreamer(id, msg)
		} else if user.Permission.IsManage {
			message = u.ToManage(id, msg)
		} else {
			message = u.ToUser(id, msg)
		}

		pushMsg, err := message.ToProto()
		if err != nil {
			return nil, err
		}

		pushMsg.Room = rid
		pushMsg.Mid = user.Id
		pushMsg.Message = msg
		pushMsg.Type = logicpb.PushMsg_ROOM
		pushMsg.MsgType = models.MESSAGE_TYPE
		pushMsg.IsRaw = true

		return pushMsg, nil
	})
}

// 發送私訊息
func (p *Producer) SendPrivate(keys []string, msg string, user *models.Member) (int64, error) {
	if err := p.rate.perSec(user.Id); err != nil {
		return 0, err
	}

	return p.Send(func(id int64) (*logicpb.PushMsg, error) {
		pushMsg, err := scheme.NewUser(*user).ToPrivate(id, msg).ToProto()
		if err != nil {
			return nil, err
		}

		pushMsg.Keys = keys
		pushMsg.Mid = user.Id
		pushMsg.Message = msg
		pushMsg.Type = logicpb.PushMsg_PUSH
		pushMsg.MsgType = models.MESSAGE_TYPE
		pushMsg.IsRaw = true

		return pushMsg, nil
	})
}

func (p *Producer) SendPrivateReply(keys []string, user *models.Member) (int64, error) {
	return p.Send(func(id int64) (*logicpb.PushMsg, error) {
		pushMsg, err := scheme.NewUser(*user).ToPrivateReply(id).ToProto()
		if err != nil {
			return nil, err
		}

		pushMsg.Keys = keys
		pushMsg.Mid = user.Id
		pushMsg.Type = logicpb.PushMsg_PUSH
		pushMsg.MsgType = models.MESSAGE_TYPE
		pushMsg.IsRaw = true

		return pushMsg, nil
	})
}

// 發送系統公告訊息
func (p *Producer) SendSystem(rid []int32, msg string) (int64, error) {
	return p.Send(func(id int64) (*logicpb.PushMsg, error) {
		return scheme.NewRoot().ToSystem(id, msg).ToRoomProto(rid)
	})
}

// 發送後台管理者訊息
func (p *Producer) SendAdmin(rid []int32, msg string) (int64, error) {
	return p.Send(func(id int64) (*logicpb.PushMsg, error) {
		u := scheme.NewRoot()

		pushMsg, err := u.ToAdmin(id, msg).ToProto()
		if err != nil {
			return nil, err
		}

		pushMsg.Room = rid
		pushMsg.Mid = u.Id
		pushMsg.Message = msg
		pushMsg.Type = logicpb.PushMsg_ROOM
		pushMsg.MsgType = models.MESSAGE_TYPE

		return pushMsg, nil
	})
}

// 發送置頂訊息
func (p *Producer) SendTop(rid []int32, msg string) (int64, error) {
	return p.Send(func(id int64) (*logicpb.PushMsg, error) {
		return scheme.NewRoot().ToTop(id, msg).ToRoomProto(rid)
	})
}

// 發送送禮訊息
func (p *Producer) SendGift(rid int32, user scheme.User, gift scheme.Gift) (int64, error) {
	return p.Send(func(id int64) (*logicpb.PushMsg, error) {
		if gift.Combo.Count == 0 {
			gift.ShowAnimation = true
		}
		return gift.ToProto(id, rid, user)
	})
}

// 發送打賞訊息
func (p *Producer) SendReward(rid int32, user scheme.User, amount, totalAmount float64) (int64, error) {
	return p.Send(func(id int64) (*logicpb.PushMsg, error) {
		return scheme.NewRewardProto(id, rid, user, amount, totalAmount)
	})
}

// 發送紅包訊息
func (p *Producer) SendRedEnvelope(rid []int32, message string, user scheme.User, redEnvelope scheme.RedEnvelope) (int64, error) {
	return p.Send(func(id int64) (*logicpb.PushMsg, error) {
		message, err := p.filterMessage(message)
		if err != nil {
			return nil, err
		}

		return redEnvelope.ToProto(id, rid, user, message)
	})
}

// 發送下注訊息
func (p *Producer) SendBets(rid []int32, user scheme.User, bet scheme.Bet) (int64, error) {
	return p.Send(func(id int64) (*logicpb.PushMsg, error) {
		return bet.ToProto(id, rid, user)
	})
}

// 發送中獎訊息
func (p *Producer) SendBetsWin(rid []int32, user scheme.User, gameName string) (int64, error) {
	return p.Send(func(id int64) (*logicpb.PushMsg, error) {
		return scheme.NewBetsWinProto(id, rid, user, gameName)
	})
}

// 發送中獎後打賞訊息
func (p *Producer) SendBetsWinReward(keys []string, user scheme.User, amount float64, buttonName string) (int64, error) {
	return p.Send(func(id int64) (*logicpb.PushMsg, error) {
		return scheme.NewBetsWinRewardProto(id, keys, user, amount, buttonName)
	})
}

// 發送等級提升提示
func (p *Producer) SendLevelUpAlert(keys []string, user *models.Member, level int) (int64, error) {
	return p.Send(func(id int64) (*logicpb.PushMsg, error) {
		pushMsg, err := scheme.LevelUpAlertMessage{
			Message: scheme.NewUser(*user).ToLevelAlert(id),
			Level:   level,
		}.ToProto()

		if err != nil {
			return nil, err
		}

		pushMsg.Keys = keys
		pushMsg.Mid = user.Id
		pushMsg.Type = logicpb.PushMsg_PUSH
		pushMsg.MsgType = models.MESSAGE_TYPE
		pushMsg.IsRaw = true

		return pushMsg, nil
	})
}

// 發送主播等級提升提示
func (p *Producer) SendAnchorLevelUpAlert(rid []int32, user *models.Member, level int) (int64, error) {
	return p.Send(func(id int64) (*logicpb.PushMsg, error) {
		pushMsg, err := scheme.LevelUpAlertMessage{ // TODO
			Message: scheme.NewUser(*user).ToAnchorLevelAlert(id),
			Level:   level,
		}.ToProto()

		if err != nil {
			return nil, err
		}

		pushMsg.Mid = user.Id
		pushMsg.Room = rid
		pushMsg.Type = logicpb.PushMsg_ROOM
		pushMsg.MsgType = models.MESSAGE_TYPE
		pushMsg.IsRaw = true

		return pushMsg, nil
	})
}

// 發送等級提升訊息
func (p *Producer) SendLevelUp(keys []string, user *models.Member, level int) (int64, error) {
	return p.Send(func(id int64) (*logicpb.PushMsg, error) {
		pushMsg, err := scheme.NewUser(*user).ToLevel(id, user, level).ToProto()
		if err != nil {
			return nil, err
		}

		pushMsg.Keys = keys
		pushMsg.Mid = user.Id
		pushMsg.Type = logicpb.PushMsg_PUSH
		pushMsg.MsgType = models.MESSAGE_TYPE
		pushMsg.IsRaw = true

		return pushMsg, nil
	})
}

// 發送進入房間訊息
func (p *Producer) SendConnect(rid int32, user *logicpb.User, isManage bool) (int64, error) {
	return p.Send(func(id int64) (*logicpb.PushMsg, error) {
		return scheme.NewConnect(id, user.Level, isManage, user.Name).ToRoomProto([]int32{rid})
	})
}

// 發送成為/移除房管訊息
func (p *Producer) SendPermission(keys []string, user *models.Member, connect logicpb.Connect) (int64, error) {
	return p.Send(func(id int64) (*logicpb.PushMsg, error) {
		if connect.Permission.IsManage {
			connect.PermissionMessage.IsManage = "你已被主播设为房管人员"
		} else {
			connect.PermissionMessage.IsManage = "你已被主播取消房管人员资格"
		}

		now := time.Now()

		msg := struct {
			scheme.Message
			Permission        logicpb.Permission        `json:"permission"`
			PermissionMessage logicpb.PermissionMessage `json:"permission_message"`
		}{
			Message: scheme.Message{
				Id:        id,
				Type:      "permission",
				Time:      now.Format("15:04:05"),
				Timestamp: now.Unix(),
			},
			Permission:        *connect.Permission,
			PermissionMessage: *connect.PermissionMessage,
		}

		bm, err := json.Marshal(msg)
		if err != nil {
			return nil, err
		}

		return &logicpb.PushMsg{
			Keys:    keys,
			Seq:     id,
			Mid:     user.Id,
			Op:      pb.OpRaw,
			Msg:     bm,
			Type:    logicpb.PushMsg_PUSH,
			MsgType: models.MESSAGE_TYPE,
			IsRaw:   true,
		}, nil
	})
}

// 發送追隨主播訊息
func (p *Producer) SendFollow(rid int32, user scheme.User, total int) (int64, error) {
	return p.Send(func(id int64) (*logicpb.PushMsg, error) {
		return scheme.NewFollowProto(id, rid, user, total)
	})
}

// TODO vip msg
// 發送一般訊息
func (p *Producer) SendMessage(rid []int32, msg scheme.Message, isRaw bool) (int64, error) {
	log.Infof("ssssssss")
	return p.Send(func(id int64) (*logicpb.PushMsg, error) {
		now := time.Now()
		msg.Id = id
		msg.Time = now.Format("15:04:05")
		msg.Timestamp = now.Unix()

		p, err := msg.ToRoomProto(rid)
		if err != nil {
			return nil, err
		}

		p.IsRaw = isRaw

		return p, nil
	})
}

// 發送禁言訊息
func (p *Producer) SendDisplay(rid []int32, user scheme.User, display scheme.Display) (int64, error) {
	return p.Send(func(id int64) (*logicpb.PushMsg, error) {
		return user.DisplayToMessage(id, display).ToRoomProto(rid)
	})
}

// 發送踢人訊息到kafka producer
func (p *Producer) Kick(msg string, keys []string) error {
	m := struct {
		Message string `json:"message"`
	}{
		Message: msg,
	}
	bm, _ := json.Marshal(m)

	pushMsg := &logicpb.PushMsg{
		Type: logicpb.PushMsg_PUSH,
		Op:   pb.OpProtoFinish,
		Keys: keys,
		Msg:  bm,
	}
	if err := p.send(pushMsg); err != nil {
		return err
	}
	return nil
}

// 發送移除置頂訊息
func (p *Producer) CloseTop(msgId int64, rid []int32) error {
	pushMsg := &logicpb.PushMsg{
		Type:  logicpb.PushMsg_ROOM,
		Op:    pb.OpCloseTopMessage,
		IsRaw: true,
		Seq:   msgId,
		Room:  rid,
		Msg:   []byte(fmt.Sprintf(`{"id":%d}`, msgId)),
	}
	if err := p.send(pushMsg); err != nil {
		return err
	}
	return nil
}

func (p *Producer) send(pushMsg *logicpb.PushMsg) error {
	b, err := proto.Marshal(pushMsg)
	if err != nil {
		return err
	}

	m := &kafka.ProducerMessage{
		Key:   kafka.StringEncoder(pushMsg.Seq),
		Topic: p.topic,
		Value: kafka.ByteEncoder(b),
	}

	// push msg to kafka
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

func (p *Producer) sends(pushMsgs []*logicpb.PushMsg) error {
	var producerMessages []*kafka.ProducerMessage
	for _, msg := range pushMsgs {
		b, err := proto.Marshal(msg)
		if err != nil {
			return err
		}

		producerMessages = append(producerMessages, &kafka.ProducerMessage{
			Key:   kafka.StringEncoder(msg.Seq),
			Topic: p.topic,
			Value: kafka.ByteEncoder(b),
		})
	}

	err := p.producer.SendMessages(producerMessages)
	if err != nil {
		log.Error(
			"message producer send messages",
			zap.Error(err),
			zap.String("topic", p.topic),
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
