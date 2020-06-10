package models

import (
	"database/sql"
	"fmt"
	"time"
)

type RoomMessage struct {
	Id     int `xorm:"pk autoincr"`
	MsgId  int64
	RoomId int
	Type   int
}

func (r *RoomMessage) TableName() string {
	return "room_messages"
}

// 頂置訊息
const TOP_MESSAGE = 1

// 公告訊息
const BULLETIN_MESSAGE = 2

type Message struct {
	Id       int `xorm:"pk autoincr"`
	MsgId    int64
	MemberId int64
	Type     string
	Message  string
	SendAt   time.Time
}

func (r *Message) TableName() string {
	return "messages"
}

type RedEnvelopeMessage struct {
	Id             int `xorm:"pk autoincr"`
	MsgId          int64
	MemberId       int64
	Message        string
	RedEnvelopesId string
	Token          string
	ExpireAt       time.Time
	SendAt         time.Time
}

func (r *RedEnvelopeMessage) TableName() string {
	return "red_envelope_messages"
}

type Messages struct {
	List               []int64
	Type               map[int64]int
	Message            map[int64]Message
	RedEnvelopeMessage map[int64]RedEnvelopeMessage
}

const (
	messageLimit = 20

	MESSAGE_TYPE      = 1
	RED_ENVELOPE_TYPE = 2
)

func (s *Store) GetRoomMessage(roomId int32, start time.Time) (*Messages, error) {
	rms := make([]RoomMessage, 0)

	err := s.d.Table(fmt.Sprintf("room_messages_%02d", roomId%50)).
		Where("`room_id` = ?", roomId).
		Where("send_at >= ? and send_at < ?", time.Now().Add(-2*time.Hour), start).
		Limit(messageLimit).
		Desc("msg_id").
		Find(&rms)
	if err != nil {
		return nil, err
	}
	if len(rms) == 0 {
		return nil, sql.ErrNoRows
	}

	var msgId []int64
	var redMsgId []int64
	msgIds := make([]int64, 0, messageLimit)
	mapMsg := make(map[int64]int)

	for i := len(rms); i > 0; i-- {
		msg := rms[i-1]
		mapMsg[msg.MsgId] = msg.Type
		msgIds = append(msgIds, msg.MsgId)

		if msg.Type == MESSAGE_TYPE {
			msgId = append(msgId, msg.MsgId)
		} else {
			redMsgId = append(redMsgId, msg.MsgId)
		}
	}

	msgs := make([]Message, 0)
	redMsgs := make([]RedEnvelopeMessage, 0)
	if err := s.d.In("msg_id", msgId).Table(fmt.Sprintf("messages_%02d", roomId%50)).Find(&msgs); err != nil {
		return nil, err
	}
	if err := s.d.In("msg_id", redMsgId).Find(&redMsgs); err != nil {
		return nil, err
	}

	msgMap := make(map[int64]Message, 0)
	redMsgMap := make(map[int64]RedEnvelopeMessage, 0)

	for _, v := range msgs {
		msgMap[v.MsgId] = v
	}
	for _, v := range redMsgs {
		redMsgMap[v.MsgId] = v
	}

	return &Messages{
		List:               msgIds,
		Type:               mapMsg,
		Message:            msgMap,
		RedEnvelopeMessage: redMsgMap,
	}, nil
}
