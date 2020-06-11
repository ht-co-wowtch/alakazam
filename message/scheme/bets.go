package scheme

import (
	"encoding/json"
	"gitlab.com/jetfueltw/cpw/alakazam/app/comet/pb"
	logicpb "gitlab.com/jetfueltw/cpw/alakazam/app/logic/pb"
	"gitlab.com/jetfueltw/cpw/alakazam/pkg/money"
	"time"
	"unicode/utf8"
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
			User:      user,
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

type BetsWinReward struct {
	Message
	Reward displayMessage `json:"reward"`
}

func NewBetsWin(seq int64, user User, gameName string) Message {
	now := time.Now()
	return Message{
		Id:        seq,
		Type:      MESSAGE_TYPE,
		User:      user,
		Display:   displayByBetsWin(user, gameName),
		Time:      now.Format("15:04:05"),
		Timestamp: now.Unix(),
	}
}

func NewBetsWinReward(seq int64, user User, amount float64, buttonName string) BetsWinReward {
	now := time.Now()
	msg := "恭喜您中奖 金额＄" + money.FormatFloat64(amount) + " "
	return BetsWinReward{
		Message: Message{
			Id:        seq,
			Type:      BETS_WIN_REWARD,
			User:      user,
			Time:      now.Format("15:04:05"),
			Timestamp: now.Unix(),
		},
		Reward: displayMessage{
			Text:            msg + buttonName,
			Color:           "#FFFFAA",
			BackgroundColor: "#F8565699",
			Entity: []textEntity{
				buttonTextEntity(buttonName, utf8.RuneCountInString(msg)),
			},
		},
	}
}

func (b BetsWinReward) ToProto(keys []string) (*logicpb.PushMsg, error) {
	bm, err := json.Marshal(b)
	if err != nil {
		return nil, err
	}

	return &logicpb.PushMsg{
		Seq:    b.Id,
		Type:   logicpb.PushMsg_PUSH,
		Op:     pb.OpRaw,
		Keys:   keys,
		Msg:    bm,
		SendAt: b.Timestamp,
	}, nil
}
