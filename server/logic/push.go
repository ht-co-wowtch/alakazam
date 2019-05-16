package logic

import (
	"encoding/json"
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
	rId, name, w, err := l.dao.UserData(p.Uid, p.Key)

	if err != nil {
		return err
	}
	if name == "" {
		return errors.LoginError
	}
	if rId == "" {
		return errors.RoomError
	}
	if l.IsBanned(p.Uid, w) {
		return errors.BannedError
	}

	msg, err := json.Marshal(Message{
		Name:    name,
		Avatar:  "",
		Message: p.Message,
		Time:    time.Now().Format("15:04:05"),
	})
	if err != nil {
		return err
	}
	return l.dao.BroadcastRoomMsg(rId, msg)
}

type PushRoomAllForm struct {
	// 要廣播的房間
	RoomId []string `form:"room_id" binding:"required"`

	// user push message
	Message string `form:"message" binding:"required"`
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
		return err
	}
	return l.dao.BroadcastMsg(p.RoomId, 0, msg)
}
