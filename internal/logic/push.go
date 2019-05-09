package logic

import (
	"context"
	"encoding/json"
)

type message struct {
	Name    string `json:"name"`
	Avatar  string `json:"avatar"`
	Message string `json:"message"`
}

type PushRoomForm struct {
	// 房間id
	RoomId string `form:"room_id" binding:"required"`

	// user uid
	Uid string `form:"uid" binding:"required"`

	// user connection key
	Key string `form:"key" binding:"required"`

	// user push message
	Message string `form:"message" binding:"required"`
}

// 單一房間推送
func (l *Logic) PushRoom(c context.Context, p *PushRoomForm) error {
	res, err := l.dao.UidInfo(p.Uid, p.Key)
	if err != nil {
		return err
	}
	msg, err := json.Marshal(message{
		Name:    res[0],
		Avatar:  "",
		Message: p.Message,
	})
	if err != nil {
		return err
	}
	return l.dao.BroadcastRoomMsg(c, p.RoomId, msg)
}

type PushRoomAllForm struct {
	// user uid
	Uid string `form:"uid" binding:"required"`

	// user connection key
	Key string `form:"key" binding:"required"`

	// user push message
	Message string `form:"message" binding:"required"`
}

// 所有房間推送
func (l *Logic) PushAll(c context.Context, p *PushRoomAllForm) error {
	res, err := l.dao.UidInfo(p.Uid, p.Key)
	if err != nil {
		return err
	}
	msg, err := json.Marshal(message{
		Name:    res[0],
		Avatar:  "",
		Message: p.Message,
	})
	if err != nil {
		return err
	}
	return l.dao.BroadcastMsg(c, 0, msg)
}
