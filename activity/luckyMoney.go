package activity

import "gitlab.com/jetfueltw/cpw/alakazam/errors"

const (
	// 普通紅包
	Money = int(1)

	// 拼手氣紅包
	LuckMoney = int(2)
)

type LuckyMoney struct {
}

func NewLuckyMoney() *LuckyMoney {
	return new(LuckyMoney)
}

// 發紅包
// TODO 未完
func (l *LuckyMoney) Give(amount float32, count int, message string, model int) (err error) {
	switch model {
	case Money:
		err = nil
	case LuckMoney:
		err = nil
	default:
		err = errors.DataError
	}

	return err
}
