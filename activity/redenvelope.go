package activity

import (
	"gitlab.com/jetfueltw/cpw/alakazam/client"
)

const (
	// 普通紅包
	Money = int(1)

	// 拼手氣紅包
	LuckMoney = int(2)
)

type Client struct {
	c *client.Client
}

func New(client *client.Client) *Client {
	return &Client{
		c: client,
	}
}

type GiveRedEnvelope struct {
	// 單包金額 or 總金額 看Type種類決定
	// 最少0.01元
	Amount int `json:"amount" binding:"required"`

	// 紅包數量,範圍1 ~ 500包
	Count int `json:"count" binding:"required"`

	// 紅包說明 下限1字元, 上限20字元
	Message string `json:"message" binding:"required,max=20"`

	// 紅包種類 拼手氣 or 普通
	Type string `json:"type" binding:"required"`

	Token string `json:"-"`
}

// 發紅包
func (l *Client) GiveRedEnvelope(roomId string, give GiveRedEnvelope) (client.RedEnvelopeReply, error) {
	r := client.RedEnvelope{
		RoomId:    roomId,
		Message:   give.Message,
		Type:      give.Type,
		Amount:    give.Amount,
		Count:     give.Count,
		ExpireMin: 60,
	}
	reply, err := l.c.GiveRedEnvelope(r, give.Token)
	if err != nil {
		return client.RedEnvelopeReply{}, err
	}
	return reply, nil
}
