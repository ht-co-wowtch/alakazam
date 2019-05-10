package logic

import (
	"context"
	"encoding/json"
	"fmt"
	"gitlab.com/jetfueltw/cpw/alakazam/server/logic/dao"
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
	if len(res) == 0 {
		return fmt.Errorf("帳號未登入")
	}
	if _, ok := res[p.Key]; !ok {
		return fmt.Errorf("沒有在房間內")
	}

	msg, err := json.Marshal(Message{
		Name:    res[dao.HashNameKey],
		Avatar:  "",
		Message: p.Message,
		Time:    time.Now().Format("15:04:05"),
	})
	if err != nil {
		return err
	}
	return l.dao.BroadcastRoomMsg(c, res[p.Key], msg)
}

type PushRoomAllForm struct {
	// 廣播者頭像
	Avatar string `json:"avatar"`

	// 要廣播的房間
	RoomId []string `form:"room_id" binding:"required"`

	// user push message
	Message string `form:"message" binding:"required"`
}

// 所有房間推送
func (l *Logic) PushAll(c context.Context, p *PushRoomAllForm) error {
	msg, err := json.Marshal(Message{
		Name:    "管理员",
		Avatar:  p.Avatar,
		Message: p.Message,
		Time:    time.Now().Format("15:04:05"),
	})
	if err != nil {
		return err
	}
	return l.dao.BroadcastMsg(c, p.RoomId, 0, msg)
}
