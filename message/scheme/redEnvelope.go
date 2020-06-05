package scheme

import (
	"encoding/json"
	"time"
)

// 紅包訊息格式
type RedEnvelopeMessage struct {
	// 訊息格式
	Message

	// 紅包資料
	RedEnvelope RedEnvelope `json:"red_envelope"`
}

func (m RedEnvelopeMessage) Score() float64 {
	return float64(m.Message.Timestamp)
}

// 紅包資料
type RedEnvelope struct {
	// 紅包id
	Id string `json:"id"`

	// 紅包token
	Token string `json:"token"`

	// 紅包多久過期
	Expired string `json:"expired"`
}

func (m RedEnvelope) MarshalBinary() (data []byte, err error) {
	return json.Marshal(m)
}

func (r RedEnvelope) ToMessage(seq int64, message string, user User) RedEnvelopeMessage {
	now := time.Now()
	return RedEnvelopeMessage{
		Message: Message{
			Id:        seq,
			Type:      RED_ENVELOPE_TYPE,
			User:      NullUser(user),
			Display:   displayByUser(user, message),
			Time:      now.Format("15:04:05"),
			Timestamp: now.Unix(),

			Uid:     user.Uid,
			Name:    user.Name,
			Avatar:  user.Avatar,
			Message: message,
		},
		RedEnvelope: r,
	}
}
