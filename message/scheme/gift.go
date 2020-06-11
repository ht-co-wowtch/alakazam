package scheme

import (
	"encoding/json"
	"gitlab.com/jetfueltw/cpw/alakazam/app/comet/pb"
	logicpb "gitlab.com/jetfueltw/cpw/alakazam/app/logic/pb"
	"time"
)

type GiftMessage struct {
	Message
	Gift Gift `json:"gift"`
}

type Gift struct {
	Id            int          `json:"gift_id"`
	Name          string       `json:"name"`
	Amount        float64      `json:"amount"`
	TotalAmount   float64      `json:"total_amount"`
	Combo         NullCombo    `json:"combo"`
	HintBox       NullHintBox  `json:"hint_box"`
	ShowAnimation bool         `json:"show_animation"`
	Message       string       `json:"message"`
	Entity        []textEntity `json:"entity"`
}

type Combo struct {
	Count      int `json:"count"`
	DurationMs int `json:"duration_ms"`
}

type NullCombo Combo

func (d NullCombo) MarshalJSON() ([]byte, error) {
	if d.Count == 0 {
		return []byte(`null`), nil
	}
	return json.Marshal(Combo(d))
}

type HintBox struct {
	DurationMs      int    `json:"duration_ms"`
	BackgroundColor string `json:"background_color"`
}

type NullHintBox HintBox

func (d NullHintBox) MarshalJSON() ([]byte, error) {
	if d.DurationMs == 0 {
		return []byte(`null`), nil
	}
	return json.Marshal(HintBox(d))
}

func (g GiftMessage) ToProto(rid int32) (*logicpb.PushMsg, error) {
	bm, err := json.Marshal(g)
	if err != nil {
		return nil, err
	}

	return &logicpb.PushMsg{
		Seq:    g.Id,
		Type:   logicpb.PushMsg_ROOM,
		Op:     pb.OpRaw,
		Room:   []int32{rid},
		Msg:    bm,
		SendAt: g.Timestamp,
		IsRaw:  true,
	}, nil
}

func (g Gift) ToMessage(seq int64, user User) GiftMessage {
	now := time.Now()
	return GiftMessage{
		Message: Message{
			Id:        seq,
			Type:      GIFT_TYPE,
			Display:   displayByGift(user, g.Name),
			User:      user,
			Time:      now.Format("15:04:05"),
			Timestamp: now.Unix(),
		},
		Gift: g,
	}
}

func NewReward(seq int64, user User, amount, totalAmount float64) GiftMessage {
	now := time.Now()
	display := displayByReward(user, amount)
	msg := display.Message.(displayMessage)
	return GiftMessage{
		Message: Message{
			Id:        seq,
			Type:      GIFT_TYPE,
			Display:   display,
			User:      user,
			Time:      now.Format("15:04:05"),
			Timestamp: now.Unix(),
		},
		Gift: Gift{
			Amount:      amount,
			TotalAmount: totalAmount,
			Message:     msg.Text,
			HintBox: NullHintBox{
				DurationMs:      3000,
				BackgroundColor: "#F8565699",
			},
			Entity: msg.Entity,
		},
	}
}
