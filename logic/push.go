package logic

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	log "github.com/golang/glog"
	"gitlab.com/jetfueltw/cpw/alakazam/client"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"gitlab.com/jetfueltw/cpw/alakazam/protocol/grpc"
	"time"
)

type Message struct {
	Uid     string `json:"uid"`
	Name    string `json:"name"`
	Avatar  string `json:"avatar"`
	Message string `json:"message"`
	Time    string `json:"time"`
}

type PushRoom struct {
	User

	// user push message
	Message string `json:"message" binding:"required,max=100"`
}

func (l *Logic) Auth(u *User) error {
	if err := l.auth(u); err != nil {
		return err
	}
	if err := l.authRoom(u); err != nil {
		return err
	}
	return nil
}

// 單一房間推送
func (l *Logic) PushRoom(c *gin.Context, p *PushRoom) error {
	if err := l.Auth(&p.User); err != nil {
		return err
	}
	if err := l.isMessage(p.RoomId, p.roomStatus, p.Uid, c.GetString("token")); err != nil {
		return err
	}

	msg, err := json.Marshal(Message{
		Uid:     p.Uid,
		Name:    p.name,
		Avatar:  "",
		Message: p.Message,
		Time:    time.Now().Format("15:04:05"),
	})
	if err != nil {
		return err
	}
	if err := l.stream.BroadcastRoomMsg(p.RoomId, msg, grpc.PushMsg_ROOM); err != nil {
		return err
	}
	return nil
}

type Money struct {
	Message

	// 紅包id
	Id string `json:"id"`

	// 紅包token
	Token string `json:"token"`

	// 紅包過期時間
	Expired int64 `json:"expired"`
}

func (l *Logic) PushRedEnvelope(give client.RedEnvelopeReply, user User) error {
	msg, err := json.Marshal(Money{
		Message: Message{
			Uid:     user.Uid,
			Name:    user.name,
			Avatar:  "",
			Message: give.Message,
			Time:    time.Now().Format("15:04:05"),
		},
		Id:      give.Uid,
		Token:   give.Token,
		Expired: give.ExpireAt.Unix(),
	})
	if err != nil {
		return err
	}
	if err := l.stream.BroadcastRoomMsg(user.RoomId, msg, grpc.PushMsg_MONEY); err != nil {
		return errors.FailureError
	}
	return nil
}

type PushRoomAllForm struct {
	// 要廣播的房間
	RoomId []string `json:"room_id" binding:"required"`

	// user push message
	Message string `json:"message" binding:"required"`
}

// 所有房間推送
func (l *Logic) PushAll(p *PushRoomAllForm) error {
	msg, err := json.Marshal(Message{
		Name:    "管理员",
		Avatar:  "",
		Message: p.Message,
		Time:    time.Now().Format("15:04:05"),
	})
	if err != nil {
		log.Errorf("pushAll json.Marshal() error(%v)", err)
		return errors.FailureError
	}
	if err := l.stream.BroadcastMsg(p.RoomId, msg); err != nil {
		return errors.FailureError
	}
	return nil
}
