package message

import (
	"database/sql"
	"github.com/go-redis/redis"
	"gitlab.com/jetfueltw/cpw/alakazam/member"
	"gitlab.com/jetfueltw/cpw/alakazam/message/scheme"
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

	mids := make([]int64, len(msg.Message)+len(msg.RedEnvelopeMessage))

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

	memberMap := make(map[int64]models.Member, len(ms))
	for _, v := range ms {
		memberMap[v.Id] = v
	}

	data := make([]interface{}, 0)
	for _, msgId := range msg.List {
		switch msg.Type[msgId] {
		case models.RED_ENVELOPE_TYPE:
			redEnvelope := msg.RedEnvelopeMessage[msgId]
			user := memberMap[redEnvelope.MemberId]
			read := scheme.RedEnvelope{
				Id:      redEnvelope.RedEnvelopesId,
				Token:   redEnvelope.Token,
				Expired: redEnvelope.ExpireAt.Format(time.RFC3339),
			}

			m := read.ToMessage(msgId, scheme.User{
				Uid:    user.Uid,
				Name:   user.Name,
				Avatar: ToAvatarName(user.Gender),
			}, redEnvelope.Message)

			m.Time = msg.Message[msgId].SendAt.Format("15:04:05")
			m.Timestamp = msg.Message[msgId].SendAt.Unix()

			data = append(data, m)
		case models.MESSAGE_TYPE:
			var m scheme.Message
			user := memberMap[msg.Message[msgId].MemberId]

			if user.Id == member.RootMid {
				m = scheme.NewRoot().ToAdmin(msgId, msg.Message[msgId].Message)
			} else {
				u := scheme.User{
					Uid:    user.Uid,
					Name:   user.Name,
					Avatar: ToAvatarName(user.Gender),
				}
				m = u.ToUser(msgId, msg.Message[msgId].Message)
			}

			m.Time = msg.Message[msgId].SendAt.Format("15:04:05")
			m.Timestamp = msg.Message[msgId].SendAt.Unix()

			data = append(data, m)
		}
	}

	if len(data) == 0 {
		return data, nil
	}
	return data, h.cache.addMessages(roomId, data)
}

type stringJson string

func (s stringJson) MarshalJSON() ([]byte, error) {
	return []byte(s), nil
}

func RoomTopMessageToMessage(msg models.RoomTopMessage) scheme.Message {
	u := scheme.NewRoot()
	message := u.ToTop(msg.MsgId, msg.Message)
	message.Time = msg.SendAt.Format("15:04:05")
	message.Timestamp = msg.SendAt.Unix()
	return message
}

func RoomBulletinMessageToMessage(msg models.RoomTopMessage) scheme.Message {
	u := scheme.NewRoot()
	message := u.ToSystem(msg.MsgId, msg.Message)
	message.Type = "bulletin"
	message.Time = msg.SendAt.Format("15:04:05")
	message.Timestamp = msg.SendAt.Unix()
	return message
}
