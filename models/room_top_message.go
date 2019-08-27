package models

import (
	"database/sql"
	"time"
)

type RoomTopMessage struct {
	RoomId  int32
	MsgId   int64
	Message string
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

func (s *Store) DeleteRoomTopMessage(msgId int64) error {
	aff, err := s.d.Where("`msg_id` = ?", msgId).Delete(&RoomTopMessage{})
	if err != nil {
		return err
	}
	if aff < 1 {
		return sql.ErrNoRows
	}
	return nil
}
