package message

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/go-xorm/xorm"
	"gitlab.com/jetfueltw/cpw/alakazam/app/logic/pb"
	"gitlab.com/jetfueltw/cpw/micro/log"
	"go.uber.org/zap"
	"runtime"
	"strconv"
	"time"
)

type MysqlConsumer struct {
	cache         *cache
	db            *xorm.EngineGroup
	ctx           context.Context
	consumer      []chan *pb.PushMsg
	consumerCount int32
}

func NewMysqlConsumer(ctx context.Context, db *xorm.EngineGroup, c *redis.Client) *MysqlConsumer {
	mysql := &MysqlConsumer{
		ctx:           ctx,
		cache:         newCache(c),
		db:            db,
		consumerCount: 5,
	}

	mysql.consumer = make([]chan *pb.PushMsg, mysql.consumerCount)
	for i := 0; i < int(mysql.consumerCount); i++ {
		mysql.consumer[i] = make(chan *pb.PushMsg, 1000)
		go mysql.run(mysql.consumer[i])
	}

	go func() {
		mysql.delCache()
		t := time.NewTicker(10 * time.Minute)
		for {
			select {
			case <-t.C:
				mysql.delCache()
			}
		}
	}()

	return mysql
}

func (m *MysqlConsumer) run(msg chan *pb.PushMsg) {
	id := goroutineID()
	for {
		select {
		case p := <-msg:
			if err := m.cache.addMessage(p); err != nil {
				log.Error("consumer message", zap.Error(messageError{
					msgId:   p.Seq,
					mid:     p.Mid,
					message: string(p.Msg),
					error:   err,
				}))
			}

			var err error
			if p.Type <= pb.PushMsg_MONEY {
				err = m.Member(p)
			} else {
				err = m.Admin(p)
			}
			if err != nil {
				log.Error("consumer message", zap.Error(err))
			}
		case <-m.ctx.Done():
			log.Infof("[goroutine %d] stop mysql consumer", id)
			return
		}
	}
}

func goroutineID() uint64 {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	n, _ := strconv.ParseUint(string(b), 10, 64)
	return n
}

const (
	addMessage             = "INSERT INTO `messages_%02d` (`msg_id`,`member_id`,`message`,`send_at`) VALUES (?,?,?,?);"
	addAdminMessage        = "INSERT INTO `admin_messages` (`msg_id`,`room_id`,`message`,`send_at`) VALUES (?,?,?,?);"
	addRedEnvelopeMessages = "INSERT INTO `red_envelope_messages` (`msg_id`,`member_id`,`message`,`red_envelopes_id`,`token`,`expire_at`,`send_at`) VALUES (?,?,?,?,?,?,?);"
	addRoomMessage         = "INSERT INTO `room_messages_%02d` (`room_id`,`msg_id`,`type`,`send_at`) VALUES (?,?,?,?);"
)

func (m *MysqlConsumer) Push(msg *pb.PushMsg) error {
	if msg.Type <= pb.PushMsg_MONEY {
		m.consumer[msg.Room[0]%m.consumerCount] <- msg
	}
	return nil
}

func (m *MysqlConsumer) Member(msg *pb.PushMsg) error {
	sendAt := time.Unix(msg.SendAt, 0)
	tx := m.db.Master().Prepare()

	defer tx.Rollback()

	switch msg.Type {
	case pb.PushMsg_MONEY:
		m := new(Money)
		if err := json.Unmarshal(msg.Msg, m); err != nil {
			return err
		}

		expireAt, err := time.Parse(time.RFC3339, m.RedEnvelope.Expired)
		if err != nil {
			log.Error("parse time for mysql consumer", zap.Error(err), zap.String("expired", m.RedEnvelope.Expired))
		}
		if _, err := tx.Exec(addRedEnvelopeMessages, msg.Seq, msg.Mid, msg.Message, m.RedEnvelope.Id, m.RedEnvelope.Token, expireAt, sendAt); err != nil {
			return &MysqlRedEnvelopeMessageError{
				error:         err,
				redEnvelopeId: m.RedEnvelope.Id,
				msgId:         msg.Seq,
				mid:           msg.Mid,
				message:       msg.Message,
			}
		}
	case pb.PushMsg_ROOM:
		if _, err := tx.Exec(fmt.Sprintf(addMessage, msg.Room[0]%50), msg.Seq, msg.Mid, msg.Message, sendAt); err != nil {
			return &messageError{
				error:   err,
				msgId:   msg.Seq,
				mid:     msg.Mid,
				message: msg.Message,
			}
		}
	}

	for _, rid := range msg.Room {
		if _, err := tx.Exec(fmt.Sprintf(addRoomMessage, rid%50), rid, msg.Seq, msg.Type, sendAt); err != nil {
			return &MysqlRoomMessageError{
				error:   err,
				msgId:   msg.Seq,
				room:    msg.Room,
				message: msg.Message,
			}
		}
	}
	return tx.Commit()
}

func (m *MysqlConsumer) Admin(msg *pb.PushMsg) error {
	sendAt := time.Unix(msg.SendAt, 0)
	tx := m.db.Master().Prepare()

	defer tx.Rollback()

	switch msg.Type {
	case pb.PushMsg_ADMIN:
		b, err := json.Marshal(msg.Room)
		if err != nil {
			return &MysqlAdminMessageError{
				error:   err,
				msgId:   msg.Seq,
				room:    msg.Room,
				message: msg.Message,
			}
		}
		if _, err := tx.Exec(addAdminMessage, msg.Seq, string(b), msg.Message, sendAt); err != nil {
			return &MysqlAdminMessageError{
				error:   err,
				msgId:   msg.Seq,
				room:    msg.Room,
				message: msg.Message,
			}
		}
	}
	return tx.Commit()
}

func (m *MysqlConsumer) delCache() {
	keys, err := m.cache.getMessageExistsKey()
	if err != nil {
		log.Error("get message exists cache key", zap.Error(err))
	} else if err := m.cache.delMessage(keys); err != nil {
		log.Error("del cache message", zap.Error(err), zap.Strings("keys", keys))
	}
}
