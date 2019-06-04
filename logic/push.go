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
	Message string `json:"message" binding:"required"`
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

	option := &client.Params{
		Uid:   p.Uid,
		Token: c.GetString("token"),
	}

	if err := l.isMessage(p.roomId, p.roomStatus, option); err != nil {
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
		log.Errorf("pushRoom json.Marshal(uid: %s ) error(%v)", p.Uid, err)
		return errors.FailureError
	}

	if err := l.stream.BroadcastRoomMsg(p.roomId, msg, grpc.PushMsg_ROOM); err != nil {
		return errors.FailureError
	}
	return nil
}

type Money struct {
	Message

	// 紅包token
	Token string `json:"token"`

	// 紅包過期時間
	Expired int64 `json:"expired"`
}

func (l *Logic) PushMoney(id, message string, user *User) error {
	msg, err := json.Marshal(Money{
		Message: Message{
			Uid:     user.Uid,
			Name:    user.name,
			Avatar:  "",
			Message: message,
			Time:    time.Now().Format("15:04:05"),
		},
		Token:   "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpYXQiOjE1NTcyMTE2NTAsIm5iZiI6MTU1NzIxMTY1MCwiaXNzIjoibG9naW4iLCJzZXNzaW9uX3Rva2VuIjoiZjc2OTYyM2Y0YTNlNDE4MWE4NzAwYWNkYTE3NzE1MmIiLCJkYXRhIjp7InVpZCI6IjEyNTdlN2Q5ZTFjOTQ0ZWY5YTZmMTI5Y2I5NDk1ZDAyIiwidXNlcm5hbWUiOiJyb290In19.7VJxH3tQpnJqWTlPbId7f0Rt7eQoaVvaJmbWxtHTqRU",
		Expired: time.Now().Add(time.Hour).Unix(),
	})

	if err != nil {
		return errors.FailureError
	}

	if err := l.stream.BroadcastRoomMsg(user.roomId, msg, grpc.PushMsg_MONEY); err != nil {
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
