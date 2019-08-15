package message

import (
	"encoding/json"
	"gitlab.com/jetfueltw/cpw/alakazam/app/logic/pb"
	"time"
)

type Message struct {
	Uid     string `json:"uid"`
	Name    string `json:"name"`
	Avatar  string `json:"avatar"`
	Message string `json:"message"`
	Time    string `json:"time"`
}

// 單一房間推送
func (l *Producer) Send(roomId string, message Message) error {
	msg, err := json.Marshal(message)
	if err != nil {
		return err
	}
	if err := l.BroadcastRoom(roomId, msg, pb.PushMsg_ROOM); err != nil {
		return err
	}
	return nil
}

type Money struct {
	Message
	RedEnvelope
}

type RedEnvelope struct {
	// 紅包id
	Id string `json:"id"`

	// 紅包token
	Token string `json:"token"`

	// 紅包過期時間
	Expired int64 `json:"expired"`
}

func (l *Producer) SendRedEnvelope(roomId string, message Message, envelope RedEnvelope) error {
	msg, err := json.Marshal(Money{
		Message:     message,
		RedEnvelope: envelope,
	})
	if err != nil {
		return err
	}
	if err := l.BroadcastRoom(roomId, msg, pb.PushMsg_MONEY); err != nil {
		return err
	}
	return nil
}

// 所有房間推送
// TODO 需實作訊息是否頂置
func (l *Producer) SendForAdmin(roomId []string, message string, isTop bool) (int64, error) {
	msg, err := json.Marshal(Message{
		Name:    "管理员",
		Message: message,
		Time:    time.Now().Format("15:04:05"),
	})
	if err != nil {
		return 0, err
	}
	var t pb.PushMsg_Type
	if isTop {
		t = pb.PushMsg_TOP
	} else {
		t = pb.PushMsg_ROOM
	}
	_, id, err := l.Broadcast(roomId, msg, t)
	if err != nil {
		return 0, err
	}
	return id, nil
}
