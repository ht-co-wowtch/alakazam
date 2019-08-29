package message

import (
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/go-xorm/xorm"
	"gitlab.com/jetfueltw/cpw/alakazam/app/logic/pb"
	"gitlab.com/jetfueltw/cpw/micro/log"
	"go.uber.org/zap"
	"time"
)

type MysqlConsumer struct {
	cache *cache
	db    *xorm.EngineGroup
}

func NewMysqlConsumer(db *xorm.EngineGroup, c *redis.Client) *MysqlConsumer {
	return &MysqlConsumer{
		cache: newCache(c),
		db:    db,
	}
}

const (
	addMessage             = "INSERT INTO `messages_%02d` (`msg_id`,`member_id`,`message`,`send_at`) VALUES (?,?,?,?);"
	addAdminMessage        = "INSERT INTO `admin_messages` (`msg_id`,`room_id`,`message`,`send_at`) VALUES (?,?,?,?);"
	addRedEnvelopeMessages = "INSERT INTO `red_envelope_messages` (`msg_id`,`member_id`,`message`,`red_envelopes_id`,`token`,`expire_at`,`send_at`) VALUES (?,?,?,?,?,?,?);"
	addRoomMessage         = "INSERT INTO `room_messages_%02d` (`room_id`,`msg_id`,`type`,`send_at`) VALUES (?,?,?,?);"
)

func (m *MysqlConsumer) Push(msg *pb.PushMsg) error {
	if msg.Type <= pb.PushMsg_MONEY {
		return m.Member(msg)
	}
	return m.Admin(msg)
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
			return &MysqlMessageError{
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
