package message

import (
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
	addMessage     = "INSERT INTO `messages` (`msg_id`,`member_id`,`message`,`send_at`) VALUES (?,?,?,?);"
	addRoomMessage = "INSERT INTO `room_messages` (`room_id`,`msg_id`,`type`,`send_at`) VALUES (?,?,?,?);"
)

func (m *MysqlConsumer) Push(msg *pb.PushMsg) error {
	sendAt := time.Unix(msg.SendAt, 0)
	tx := m.db.Master().Prepare()

	defer tx.Rollback()
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
	for _, rid := range msg.Rids {
		if _, err := tx.Exec(addRoomMessage, rid, msg.Seq, msg.Type, sendAt); err != nil {
			log.Error(
				"insert room message",
				zap.Error(err),
				zap.Int64("room", rid),
				zap.Int64("msg_id", msg.Seq),
			)
			return err
		}
	}
	return tx.Commit()
}
