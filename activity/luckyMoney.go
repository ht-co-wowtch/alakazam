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

type RedEnvelopeClient struct {
	client *client.Client
}

func NewRedEnvelope(client *client.Client) *RedEnvelopeClient {
	return &RedEnvelopeClient{
		client: client,
	}
}

type GiveRedEnvelope struct {
	// 單包金額 or 總金額 看Type種類決定
	// 最少0.01元
	Amount float64 `json:"amount" binding:"required,min=0.01"`

	// 紅包數量,範圍1 ~ 500包
	Count int `json:"count" binding:"required,min=1,max=500"`

	// 紅包說明 下限1字元, 上限20字元
	Message string `json:"message" binding:"required,min=1,max=20"`

	// 紅包種類 拼手氣 or 普通
	Type int `json:"type" binding:"required"`

	Token string `json:"-"`
}

// 發紅包
// TODO 未完
func (l *RedEnvelopeClient) Give(money *GiveRedEnvelope) (string, error) {
	return "", nil
}
