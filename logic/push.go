package logic

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	log "github.com/golang/glog"
	"gitlab.com/jetfueltw/cpw/alakazam/client"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"time"
)

type Message struct {
	Uid     string `json:"uid"`
	Name    string `json:"name"`
	Avatar  string `json:"avatar"`
	Message string `json:"message"`
	Time    string `json:"time"`
}

type PushRoomJson struct {
	User

	// user push message
	Message string `json:"message" binding:"required"`
}

// 單一房間推送
func (l *Logic) PushRoom(c *gin.Context, p *PushRoomJson) error {
	if err := l.auth(&p.User); err != nil {
		return err
	}
	if err := l.authRoom(&p.User); err != nil {
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

	if err := l.stream.BroadcastRoomMsg(p.roomId, msg); err != nil {
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
