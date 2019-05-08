package logic

import (
	"context"
	"encoding/json"
)

// 單一房間推送
func (l *Logic) PushRoom(c context.Context, room string, msg []byte) (err error) {
	if msg, err = l.toMessage(msg); err == nil {
		return l.dao.BroadcastRoomMsg(c, room, msg)
	}
	return
}

// 所有房間推送
func (l *Logic) PushAll(c context.Context, speed int32, msg []byte) (err error) {
	if msg, err = l.toMessage(msg); err == nil {
		return l.dao.BroadcastMsg(c, speed, msg)
	}
	return
}

type message struct {
	Name    string `json:"name"`
	Avatar  string `json:"avatar"`
	Message string `json:"message"`
}

func (l *Logic) toMessage(msg []byte) ([]byte, error) {
	m := message{
		Name:    "",
		Avatar:  "",
		Message: string(msg),
	}
	return json.Marshal(m)
}
