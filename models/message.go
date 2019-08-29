package models

import (
	"fmt"
	"gitlab.com/jetfueltw/cpw/alakazam/app/logic/pb"
	"time"
)

type RoomMessage struct {
	Id     int `xorm:"pk autoincr"`
	MsgId  int64
	RoomId int
	Type   pb.PushMsg_Type
}

func (r *RoomMessage) TableName() string {
	return "room_messages"
}

type Message struct {
	Id       int `xorm:"pk autoincr"`
	MsgId    int64
	MemberId int
	Message  string
	SendAt   time.Time
}

func (r *Message) TableName() string {
	return "messages"
}

type RedEnvelopeMessage struct {
	Id             int `xorm:"pk autoincr"`
	MsgId          int64
	MemberId       int
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
	Type               map[int64]pb.PushMsg_Type
	Message            map[int64]Message
	RedEnvelopeMessage map[int64]RedEnvelopeMessage
}

const (
	messageLimit = 20
)

func (s *Store) GetRoomMessage(roomId, lastMsgId int) (*Messages, error) {
	rms := make([]RoomMessage, 0)

	query := s.d.Table(fmt.Sprintf("room_messages_%02d", roomId%50)).
		Where("`room_id` = ?", roomId)

	if lastMsgId > 0 {
		query = query.Where("`msg_id` < ?", lastMsgId)
	}

	err := query.Limit(messageLimit).
		Desc("msg_id").
		Find(&rms)
	if err != nil {
		return nil, err
	}

	var msgId []int64
	var redMsgId []int64
	msgIds := make([]int64, 0, messageLimit)
	mapMsg := make(map[int64]pb.PushMsg_Type)

	for i := len(rms); i > 0; i-- {
		msg := rms[i-1]
		mapMsg[msg.MsgId] = msg.Type
		msgIds = append(msgIds, msg.MsgId)
		switch msg.Type {
		case pb.PushMsg_MONEY:
			redMsgId = append(redMsgId, msg.MsgId)
		default:
			msgId = append(msgId, msg.MsgId)
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
