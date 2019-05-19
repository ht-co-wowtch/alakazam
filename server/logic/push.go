package logic

import (
	"encoding/json"
	log "github.com/golang/glog"
	"gitlab.com/jetfueltw/cpw/alakazam/server/errors"
	"time"
)

type Message struct {
	Name    string `json:"name"`
	Avatar  string `json:"avatar"`
	Message string `json:"message"`
	Time    string `json:"time"`
}

type PushRoomForm struct {
	// user uid
	Uid string `json:"uid" binding:"required"`

	// user connection key
	Key string `json:"key" binding:"required"`

	// user push message
	Message string `json:"message" binding:"required"`
}

// 單一房間推送
func (l *Logic) PushRoom(p *PushRoomForm) error {
	rId, name, w, err := l.cache.GetUser(p.Uid, p.Key)
	if err != nil {
		return errors.FailureError
	}
	if name == "" {
		return errors.LoginError
	}
	if rId == "" {
		return errors.RoomError
	}
	if l.isBanned(p.Uid, w) {
		return errors.BannedError
	}

	msg, err := json.Marshal(Message{
		Name:    name,
		Avatar:  "",
		Message: p.Message,
		Time:    time.Now().Format("15:04:05"),
	})
	if err != nil {
		log.Errorf("pushRoom json.Marshal(uid: %s ) error(%v)", p.Uid, err)
		return errors.FailureError
	}
	if err := l.dao.BroadcastRoomMsg(rId, msg); err != nil {
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
	if err := l.dao.BroadcastMsg(p.RoomId, msg); err != nil {
		return errors.FailureError
	}
	return nil
}
