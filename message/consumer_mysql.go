package message

import (
	"encoding/json"
	"github.com/go-xorm/xorm"
	"gitlab.com/jetfueltw/cpw/alakazam/app/logic/pb"
	"gitlab.com/jetfueltw/cpw/micro/log"
	"go.uber.org/zap"
	"time"
)

type MysqlConsumer struct {
	db *xorm.EngineGroup
}

func NewMysqlConsumer(db *xorm.EngineGroup) *MysqlConsumer {
	return &MysqlConsumer{
		db: db,
	}
}

const (
	addMessage             = "INSERT INTO `messages` (`msg_id`,`member_id`,`message`,`send_at`) VALUES (?,?,?,?);"
	addRedEnvelopeMessages = "INSERT INTO `red_envelope_messages` (`msg_id`,`member_id`,`message`,`red_envelopes_id`,`token`,`expire_at`,`send_at`) VALUES (?,?,?,?,?,?,?);"
	addRoomMessage         = "INSERT INTO `room_messages` (`room_id`,`msg_id`,`type`,`send_at`) VALUES (?,?,?,?);"
)

func (m *MysqlConsumer) Push(msg *pb.PushMsg) error {
	sendAt := time.Unix(msg.SendAt, 0)
	tx := m.db.Master().Prepare()

	defer tx.Rollback()

	switch msg.Type {
	case pb.PushMsg_MONEY:
		m := new(Money)
		if err := json.Unmarshal(msg.Msg, m); err != nil {
			return err
		}
		expireAt := time.Unix(m.RedEnvelope.Expired, 0)
		if _, err := tx.Exec(addRedEnvelopeMessages, msg.Seq, msg.Mid, msg.Message, m.RedEnvelope.Id, m.RedEnvelope.Token, expireAt, sendAt); err != nil {
			log.Error(
				"insert red envelope message",
				zap.Error(err),
				zap.Int64("msg_id", msg.Seq),
				zap.Int64("member_id", msg.Mid),
				zap.String("message", msg.Message),
				zap.String("red envelope id", m.RedEnvelope.Id),
			)
			return err
		}
	case pb.PushMsg_ROOM, pb.PushMsg_TOP:
		if _, err := tx.Exec(addMessage, msg.Seq, msg.Mid, msg.Message, sendAt); err != nil {
			log.Error(
				"insert message",
				zap.Error(err),
				zap.Int64("msg_id", msg.Seq),
				zap.Int64("member_id", msg.Mid),
				zap.String("message", msg.Message),
			)
			return err
		}
	}

	if msg.Type != pb.PushMsg_Close {
		for _, rid := range msg.Room {
			if _, err := tx.Exec(addRoomMessage, rid, msg.Seq, msg.Type, sendAt); err != nil {
				log.Error(
					"insert room message",
					zap.Error(err),
					zap.Int32("room", rid),
					zap.Int64("msg_id", msg.Seq),
				)
				return err
			}
		}
		return tx.Commit()
	}
	return nil
}
