package logic

import (
	"context"
)

// 單一房間推送
func (l *Logic) PushRoom(c context.Context, room string, msg []byte) (err error) {
	return l.dao.BroadcastRoomMsg(c, room, msg)
}

// 所有房間推送
func (l *Logic) PushAll(c context.Context, speed int32, msg []byte) (err error) {
	return l.dao.BroadcastMsg(c, speed, msg)
}
