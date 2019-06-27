package client

import (
	"encoding/json"
	"time"
)

type Money struct {
	Dml     int `json:"dml"`
	Deposit int `json:"deposit"`
}

// TODO 待實作
func (c *Client) GetDepositAndDml(day int, uid, token string) (money Money, err error) {
	return money, err
}

type RedEnvelope struct {
	// 房間id
	RoomId string `json:"room_id"`

	// 紅包訊息
	Message string `json:"message"`

	// 紅包種類
	Type string `json:"type"`

	// 紅包金額 看種類決定
	Amount int `json:"amount"`

	// 紅包數量
	Count int `json:"count"`

	// 紅包多久過期(分鐘)
	ExpireMin int `json:"expire_min"`
}

type RedEnvelopeReply struct {
	// 訂單id
	Order string `json:"id"`

	// 誰發的紅包
	Uid string `json:"uid"`

	// 紅包種類
	Type string `json:"type"`

	// 紅包訊息
	Message string `json:"message"`

	// 紅包總金額
	TotalAmount int `json:"total_amount"`

	// 紅包數量
	TotalCount int `json:"count"`

	// 預計發送時間
	PublishAt time.Time `json:"publish_at"`

	// 過期時間
	ExpireAt time.Time `json:"expire_at"`

	// 紅包token
	Token string `json:"token"`
}

func (c *Client) GiveRedEnvelope(envelope RedEnvelope, token string) (RedEnvelopeReply, error) {
	resp, err := c.c.PostJson("/red-envelope", nil, envelope, bearer(token))
	if err != nil {
		return RedEnvelopeReply{}, err
	}
	defer resp.Body.Close()
	if err := checkResponse(resp); err != nil {
		return RedEnvelopeReply{}, err
	}
	var u RedEnvelopeReply
	if err := json.NewDecoder(resp.Body).Decode(&u); err != nil {
		return RedEnvelopeReply{}, err
	}
	return u, nil
}
