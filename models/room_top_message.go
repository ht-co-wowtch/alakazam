package models

import (
	"database/sql"
	"github.com/go-xorm/xorm"
	"time"
)

const (
	addRoomTopMessage = "REPLACE INTO `room_top_messages` (`room_id`,`msg_id`,`message`,`type`,`send_at`) VALUES (?,?,?,?,?);"
)

type RoomTopMessage struct {
	RoomId  int32
	MsgId   int64
	Message string
	Type    int
	SendAt  time.Time
}

func (r *RoomTopMessage) TableName() string {
	return "room_top_messages"
}

func (s *Store) FindRoomTopMessage(msgId int64) ([]RoomTopMessage, error) {
	msg := make([]RoomTopMessage, 0)
	err := s.d.Where("`msg_id` = ?", msgId).Find(&msg)
	return msg, err
}

func (s *Store) AddRoomTopMessage(rids []int32, message RoomTopMessage) error {
	_, err := s.d.Transaction(func(session *xorm.Session) (i interface{}, e error) {
		for _, rid := range rids {
			if _, err := session.Exec(addRoomTopMessage, rid, message.MsgId, message.Message, message.Type, message.SendAt); err != nil {
				return nil, err
			}
		}
		return nil, nil
	})
	return err
}

func (s *Store) DeleteRoomTopMessage(rid []int32, msgId int64, t int) error {
	aff, err := s.d.In("room_id", rid).
		Where("`type` = ? AND `msg_id` = ? ", t, msgId).
		Delete(&RoomTopMessage{})
	if err != nil {
		return err
	}
	if aff < 1 {
		return sql.ErrNoRows
	}
	return nil
}
