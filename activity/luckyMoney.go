package activity

import (
	"github.com/google/uuid"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
)

const (
	// 普通紅包
	Money = int(1)

	// 拼手氣紅包
	LuckMoney = int(2)
)

type moneyApi interface {
	NewOlder(id string, total float32, token string) error
}

type storeApi struct {
}

type queueApi struct {
}

type LuckyMoney struct {
	money moneyApi

	store storeApi

	queue queueApi
}

func NewLuckyMoney() *LuckyMoney {
	return new(LuckyMoney)
}

type GiveMoney struct {
	// 單包金額 or 總金額 看Type種類決定
	// 最少0.01元
	Amount float32 `json:"amount" binding:"required,min=0.01"`

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
func (l *LuckyMoney) Give(money *GiveMoney) (err error) {
	var total float32
	id := uuid.New().String()

	switch money.Type {
	case Money:
		total = float32(money.Count) * money.Amount
	case LuckMoney:
		err = nil
	default:
		return errors.DataError
	}

	if err := l.money.NewOlder(id, total, money.Token); err != nil {
		return err
	}

	return err
}
