package activity

import (
	"fmt"
	"github.com/google/uuid"
	"gitlab.com/jetfueltw/cpw/alakazam/client"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"strconv"
)

const (
	// 普通紅包
	Money = int(1)

	// 拼手氣紅包
	LuckMoney = int(2)
)

type moneyApi interface {
	NewOlder(older client.Older, option *client.Params) (float64, error)
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

func NewLuckyMoney(money moneyApi) *LuckyMoney {
	return &LuckyMoney{
		money: money,
	}
}

type GiveMoney struct {
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
func (l *LuckyMoney) Give(money *GiveMoney) (string, error) {
	var total float64

	amount, err := strconv.ParseFloat(fmt.Sprintf("%.2f", money.Amount), 64)

	if err != nil || money.Amount != amount {
		return "", errors.AmountError
	}

	switch money.Type {
	case Money:
		total = float64(money.Count) * money.Amount
	case LuckMoney:
		total = money.Amount
	default:
		return "", errors.DataError
	}

	c := client.Older{
		OrderId: uuid.New().String(),
		Amount:  total,
	}

	p := &client.Params{
		Token: money.Token,
	}

	if _, err := l.money.NewOlder(c, p); err != nil {
		switch err {
		case client.InsufficientBalanceError:
			return "", errors.BalanceError
		default:
			return "", errors.FailureError
		}
	}

	return c.OrderId, nil
}
