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
	addMessage = "INSERT INTO `messages` (`seq`,`member_id`,`room_id`,`type`,`message`,`send_at`) VALUES (?,?,?,?,?,?);"
)

func (m *MysqlConsumer) Push(msg *pb.PushMsg) error {
	sendAt := time.Unix(msg.SendAt, 0)
	for _, rid := range msg.Rids {
		_, err := m.db.Master().Exec(addMessage, msg.Seq, msg.Mid, rid, msg.Type, msg.Message, sendAt)
		if err != nil {
			log.Error("insert message", zap.Error(err), zap.Int64("room", rid), zap.Int64("seq", msg.Seq))
		}
	}
	return nil
}
