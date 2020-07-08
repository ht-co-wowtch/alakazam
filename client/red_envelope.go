package client

import (
	"encoding/json"
	"time"
)

type RedEnvelope struct {
	// 房間id
	RoomId int `json:"room_id"`

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

type RedEnvelopeAdmin struct {
	RedEnvelope

	// 什麼時候發佈
	PublishAt time.Time `json:"publish_at"`
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

	// 紅包是否來自後台
	IsAdmin bool `json:"is_admin"`

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

func (c *Client) GiveRedEnvelopeForAdmin(envelope RedEnvelopeAdmin) (RedEnvelopeReply, error) {
	resp, err := c.c.PostJson("/admin/red-envelope", nil, envelope, nil)
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

func (c *Client) UpRedEnvelopePublish(envelopeId string) error {
	publish := struct {
		Order string `json:"order"`
	}{
		Order: envelopeId,
	}
	resp, err := c.c.PostJson("/admin/red-envelope/publish", nil, publish, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if err := checkResponse(resp); err != nil {
		return err
	}
	return nil
}

type TakeEnvelopeReply struct {
	// 紅包id
	Id string `json:"id"`

	// 發紅包人的uid
	Uid string `json:"uid"`

	// 紅包訊息
	Message string `json:"message"`

	// 搶紅包的狀態
	Status string `json:"status"`

	// 搶紅包的狀態說明
	StatusMessage string `json:"status_message"`

	// 搶到的金額
	Amount float64 `json:"amount"`

	// 紅包是否來自後台
	IsAdmin bool `json:"is_admin"`
}

const (
	// 搶紅包成功
	TakeEnvelopeSuccess = "success"

	// 已經搶過該紅包
	TakeEnvelopeReceived = "received"

	// 紅包已搶完
	TakeEnvelopeGone = "gone"

	// 紅包已過期
	TakeEnvelopeExpired = "expired"
)

// 搶紅包
func (c *Client) TakeRedEnvelope(redEnvelopeToken, token string) (TakeEnvelopeReply, error) {
	resp, err := c.c.PutJson("/red-envelope", nil, map[string]string{"token": redEnvelopeToken}, bearer(token))
	if err != nil {
		return TakeEnvelopeReply{}, err
	}
	defer resp.Body.Close()
	if err := checkResponse(resp); err != nil {
		return TakeEnvelopeReply{}, err
	}
	var u TakeEnvelopeReply
	if err := json.NewDecoder(resp.Body).Decode(&u); err != nil {
		return TakeEnvelopeReply{}, err
	}
	return u, nil
}

type RedEnvelopeDetail struct {
	RedEnvelopeInfo

	// 哪些會員搶走
	Members []MemberDetail `json:"members"`
}

type RedEnvelopeInfo struct {
	// 紅包id
	Id string `json:"id"`

	// 發紅包的會員uid
	Uid string `json:"uid"`

	// 紅包訊息
	Message string `json:"message"`

	// 搶到的金額
	Amount float64 `json:"amount"`

	// 紅包總數
	Count int `json:"count"`

	// 紅包種類
	Type string `json:"type"`

	// 總金額
	TotalAmount int `json:"total_amount"`

	// 已拿走包數
	TakeCount int `json:"take_count"`

	// 已拿走金額
	TakeAmount float64 `json:"take_amount"`

	// 紅包過期時間
	ExpireAt time.Time `json:"expire_at"`

	// 紅包是否來自後台
	IsAdmin bool `json:"is_admin"`
}

type MemberDetail struct {
	// 搶走紅包會員uid
	Uid string `json:"uid"`

	// 搶走紅包會員拿走多少金額
	Amount float64 `json:"amount"`

	// 搶走紅包的時間
	TakeAt time.Time `json:"take_at"`
}

// 取紅包明細
func (c *Client) GetRedEnvelopeDetail(id, token string) (RedEnvelopeDetail, error) {
	resp, err := c.c.Get("/red-envelope/"+id, nil, bearer(token))
	if err != nil {
		return RedEnvelopeDetail{}, err
	}
	defer resp.Body.Close()
	if err := checkResponse(resp); err != nil {
		return RedEnvelopeDetail{}, err
	}
	var u RedEnvelopeDetail
	if err := json.NewDecoder(resp.Body).Decode(&u); err != nil {
		return RedEnvelopeDetail{}, err
	}
	return u, nil
}
