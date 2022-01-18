package scheme

import (
	"encoding/json"
	"gitlab.com/ht-co/wowtch/live/alakazam/app/comet/pb"
	logicpb "gitlab.com/ht-co/wowtch/live/alakazam/app/logic/pb"
	"gitlab.com/ht-co/wowtch/live/alakazam/models"
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

func (r RedEnvelope) ToMessage(seq int64, user User, message string) RedEnvelopeMessage {
	now := time.Now()
	return RedEnvelopeMessage{
		Message: Message{
			Id:        seq,
			Type:      RED_ENVELOPE_TYPE,
			User:      user,
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

func (r RedEnvelope) ToProto(seq int64, rid []int32, user User, message string) (*logicpb.PushMsg, error) {
	m := r.ToMessage(seq, user, message)
	bm, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	return &logicpb.PushMsg{
		Seq:     seq,
		Type:    logicpb.PushMsg_ROOM,
		Op:      pb.OpRaw,
		Room:    rid,
		Mid:     user.Id,
		Msg:     bm,
		MsgType: models.RED_ENVELOPE_TYPE,
		Message: message,
		SendAt:  m.Timestamp,
	}, nil
}
