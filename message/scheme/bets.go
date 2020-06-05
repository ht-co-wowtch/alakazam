package scheme

import (
	"time"
)

// 跟投訊息格式
type Bets struct {
	Message

	// 跟投資料
	Bet Bet `json:"bet"`

	// TODO 以下待廢棄
	GameId       int        `json:"game_id"`
	PeriodNumber int        `json:"period_number"`
	Items        []BetOrder `json:"bets"`
	Count        int        `json:"count"`
	TotalAmount  int        `json:"total_amount"`
}

// 跟投資料
type Bet struct {
	GameId       int        `json:"game_id"`
	GameName     string     `json:"game_name"`
	PeriodNumber int        `json:"period_number"`
	Count        int        `json:"count"`
	TotalAmount  int        `json:"total_amount"`
	Orders       []BetOrder `json:"bets"`
}

// 跟投項目資料
type BetOrder struct {
	Name       string   `json:"name"`
	OddsCode   string   `json:"odds_code"`
	Items      []string `json:"items"`
	TransItems []string `json:"trans_items"`
	Amount     int      `json:"amount"`
}

func (b Bet) ToMessage(seq int64, user User) Bets {
	now := time.Now()

	// 避免Items與TransItems欄位json Marshal後出現null
	for i, v := range b.Orders {
		if len(v.Items) == 0 {
			b.Orders[i].Items = []string{}
		}
		if len(v.TransItems) == 0 {
			b.Orders[i].TransItems = []string{}
		}
	}

	return Bets{
		Message: Message{
			Id:        seq,
			Type:      "bets",
			Display:   displayByBets(user, b.GameName, b.TotalAmount),
			User:      NullUser(user),
			Time:      now.Format("15:04:05"),
			Timestamp: now.Unix(),

			Uid:    user.Uid,
			Name:   user.Name,
			Avatar: user.Avatar,
		},

		Bet: b,

		GameId:       b.GameId,
		PeriodNumber: b.PeriodNumber,
		Items:        b.Orders,
		Count:        b.Count,
		TotalAmount:  b.TotalAmount,
	}
}

func NewBetsPay(seq int64, user User, gameName string) Message {
	now := time.Now()
	return Message{
		Id:        seq,
		Type:      MESSAGE_TYPE,
		User:      NullUser(user),
		Display:   displayByBetsPay(user, gameName),
		Time:      now.Format("15:04:05"),
		Timestamp: now.Unix(),
	}
}
