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
			if !p.IsSave {
				continue
			}

			if err := m.cache.addMessage(p); err != nil {
				log.Error("consumer message", zap.Error(messageError{
					msgId:   p.Seq,
					mid:     p.Mid,
					message: string(p.Msg),
					error:   err,
				}))
			}

			sendAt := time.Unix(p.SendAt, 0)
			tx := m.db.Master().Prepare()

			var err error

			defer tx.Rollback()

			switch p.Type {
			case pb.PushMsg_USER, pb.PushMsg_ADMIN:
				if _, e := tx.Exec(fmt.Sprintf(addMessage, p.Room[0]%50), p.Seq, p.Mid, p.Type, p.Message, sendAt); e != nil {
					err = &messageError{
						error:   e,
						msgId:   p.Seq,
						mid:     p.Mid,
						message: p.Message,
					}
				}
			case pb.PushMsg_MONEY:
				m := new(RedEnvelopeMessage)
				if err = json.Unmarshal(p.Msg, m); err != nil {
					break
				}

				expireAt, e := time.Parse(time.RFC3339, m.RedEnvelope.Expired)
				if e != nil {
					err = fmt.Errorf("parse time for mysql consumer error: %s expired: %s", e.Error(), m.RedEnvelope.Expired)
					break
				}
				if _, e := tx.Exec(addRedEnvelopeMessages, p.Seq, p.Mid, p.Message, m.RedEnvelope.Id, m.RedEnvelope.Token, expireAt, sendAt); e != nil {
					err = &MysqlRedEnvelopeMessageError{
						error:         e,
						redEnvelopeId: m.RedEnvelope.Id,
						msgId:         p.Seq,
						mid:           p.Mid,
						message:       p.Message,
					}
				}
			default:
				continue
			}

			if err == nil {
				for _, rid := range p.Room {
					if _, e := tx.Exec(fmt.Sprintf(addRoomMessage, rid%50), rid, p.Seq, p.Type, sendAt); e != nil {
						err = &MysqlRoomMessageError{
							error:   e,
							msgId:   p.Seq,
							room:    p.Room,
							message: p.Message,
						}
						break
					}
				}

				if err == nil {
					err = tx.Commit()
				}
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
	addMessage             = "INSERT INTO `messages_%02d` (`msg_id`,`member_id`,`type`,`message`,`send_at`) VALUES (?,?,?,?,?);"
	addRedEnvelopeMessages = "INSERT INTO `red_envelope_messages` (`msg_id`,`member_id`,`message`,`red_envelopes_id`,`token`,`expire_at`,`send_at`) VALUES (?,?,?,?,?,?,?);"
	addRoomMessage         = "INSERT INTO `room_messages_%02d` (`room_id`,`msg_id`,`type`,`send_at`) VALUES (?,?,?,?);"
)

func (m *MysqlConsumer) Push(msg *pb.PushMsg) error {
	if msg.Type <= pb.PushMsg_MONEY {
		m.consumer[msg.Room[0]%m.consumerCount] <- msg
	}
	return nil
}

func (m *MysqlConsumer) delCache() {
	keys, err := m.cache.getMessageExistsKey()
	if err != nil {
		log.Error("get message exists cache key", zap.Error(err))
	} else if err := m.cache.delMessage(keys); err != nil {
		log.Error("del cache message", zap.Error(err), zap.Strings("keys", keys))
	}
}
