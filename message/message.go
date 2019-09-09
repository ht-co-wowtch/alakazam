package message

import (
	"database/sql"
	"encoding/json"
	"github.com/go-redis/redis"
	"gitlab.com/jetfueltw/cpw/alakazam/app/logic/pb"
	"gitlab.com/jetfueltw/cpw/alakazam/member"
	"gitlab.com/jetfueltw/cpw/alakazam/models"
	"time"
)

type History struct {
	db     *models.Store
	member *member.Member
	cache  Cache
}

func NewHistory(db *models.Store, c *redis.Client, member *member.Member) *History {
	return &History{
		db:     db,
		cache:  newCache(c),
		member: member,
	}
}

const (
	// 普通訊息
	messageType = "message"
	// 紅包訊息
	redEnvelopeType = "red_envelope"
	// 公告訊息
	TopType = "top"
)

func (h *History) Get(roomId int32, at time.Time) ([]interface{}, error) {
	if time.Now().Add(-2 * time.Hour).After(at) {
		return []interface{}{}, nil
	}

	msgs, err := h.cache.getMessage(roomId, at)
	if err != nil {
		return []interface{}{}, nil
	}

	if len(msgs) > 0 {
		message := make([]interface{}, 0, len(msgs))
		for i := len(msgs); i > 0; i-- {
			b := msgs[i-1]
			message = append(message, stringJson(b))
		}
		return message, nil
	}

	msg, err := h.db.GetRoomMessage(roomId, at)
	if err != nil {
		if err != sql.ErrNoRows {
			return nil, err
		}
		return []interface{}{}, nil
	}

	mids := make([]int, len(msg.Message)+len(msg.RedEnvelopeMessage))

	for _, v := range msg.Message {
		mids = append(mids, v.MemberId)
	}
	for _, v := range msg.RedEnvelopeMessage {
		mids = append(mids, v.MemberId)
	}

	ms, err := h.member.GetMembers(mids)
	if err != nil {
		return []interface{}{}, err
	}

	memberMap := make(map[int]models.Member, len(ms))
	for _, v := range ms {
		memberMap[v.Id] = v
	}

	data := make([]interface{}, 0)
	for _, msgId := range msg.List {
		switch msg.Type[msgId] {
		case pb.PushMsg_MONEY:
			redEnvelope := msg.RedEnvelopeMessage[msgId]
			user := memberMap[redEnvelope.MemberId]
			data = append(data, RedEnvelopeMessage{
				Message: Message{
					Id:        msgId,
					Uid:       user.Uid,
					Name:      user.Name,
					Type:      redEnvelopeType,
					Avatar:    toAvatarName(user.Gender),
					Message:   redEnvelope.Message,
					Time:      redEnvelope.SendAt.Format("15:04:05"),
					Timestamp: redEnvelope.SendAt.Unix(),
				},
				RedEnvelope: RedEnvelope{
					Id:      redEnvelope.RedEnvelopesId,
					Token:   redEnvelope.Token,
					Expired: redEnvelope.ExpireAt.Format(time.RFC3339),
				},
			})
		case pb.PushMsg_ROOM:
			user := memberMap[msg.Message[msgId].MemberId]
			data = append(data, Message{
				Id:        msgId,
				Uid:       user.Uid,
				Name:      user.Name,
				Type:      messageType,
				Avatar:    toAvatarName(user.Gender),
				Message:   msg.Message[msgId].Message,
				Time:      msg.Message[msgId].SendAt.Format("15:04:05"),
				Timestamp: msg.Message[msgId].SendAt.Unix(),
			})
		}
	}
	return data, h.cache.addMessages(roomId, data)
}

type stringJson string

func (s stringJson) MarshalJSON() ([]byte, error) {
	return []byte(s), nil
}

func RoomTopMessageToMessage(msg models.RoomTopMessage) Message {
	return Message{
		Id:        msg.MsgId,
		Uid:       RootUid,
		Type:      TopType,
		Name:      RootName,
		Message:   msg.Message,
		Time:      msg.SendAt.Format("15:04:05"),
		Timestamp: msg.SendAt.Unix(),
	}
}

func ToMessage(msgByte []byte) (Message, error) {
	var msg Message
	err := json.Unmarshal(msgByte, &msg)
	return msg, err
}
