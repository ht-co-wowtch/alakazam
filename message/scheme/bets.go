package scheme

import (
	"encoding/json"
	"gitlab.com/ht-co/wowtch/live/alakazam/app/comet/pb"
	logicpb "gitlab.com/ht-co/wowtch/live/alakazam/app/logic/pb"
	"gitlab.com/ht-co/wowtch/live/alakazam/pkg/money"
	"time"
	"unicode/utf8"
)

// Bets 跟投訊息格式
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

// Bet 跟投資料
type Bet struct {
	GameId       int        `json:"game_id"`
	GameName     string     `json:"game_name"`
	PeriodNumber int        `json:"period_number"`
	Count        int        `json:"count"`
	TotalAmount  int        `json:"total_amount"`
	Orders       []BetOrder `json:"bets"`
}

// BetOrder 跟投項目資料
type BetOrder struct {
	Name       string   `json:"name"`
	OddsCode   string   `json:"odds_code"`
	Items      []string `json:"items"`
	TransItems []string `json:"trans_items"`
	Amount     int      `json:"amount"`
}

func (b Bet) ToProto(seq int64, rid []int32, user User) (*logicpb.PushMsg, error) {
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

	m := Bets{
		Message: Message{
			Id:        seq,
			Type:      BETS,
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

	bm, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	return &logicpb.PushMsg{
		Seq:    seq,
		Op:     pb.OpRaw,
		Type:   logicpb.PushMsg_ROOM,
		Room:   rid,
		Msg:    bm,
		SendAt: m.Timestamp,
	}, nil
}

// QuizBets 競猜跟投格式
type QuizBets struct {
	Message

	QuizBet QuizBet `json:"quiz_bet"`
}

type QuizBet struct {
	QuizId   int    `json:"quiz_id"`
	QuizName string `json:"quiz_name"`
	BetName  string `json:"bet_name"`
	Bet      string `json:"bet"`
	Amount   int    `json:"amount"`
}

// ToProto
//TODO
func (b QuizBet) ToProto(seq int64, rid []int32, user User) (*logicpb.PushMsg, error) {
	now := time.Now()
	m := QuizBets{
		Message: Message{
			Id:        seq,
			Type:      QUIZ_BETS,
			Display:   displayByBets(user, b.QuizName, b.Amount),
			User:      user,
			Time:      now.Format("15:04:05"),
			Timestamp: now.Unix(),

			Uid:    user.Uid,
			Name:   user.Name,
			Avatar: user.Avatar,
		},
		QuizBet: b,
	}

	bm, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	return &logicpb.PushMsg{
		Seq:    seq,
		Op:     pb.OpRaw,
		Type:   logicpb.PushMsg_ROOM,
		Room:   rid,
		Msg:    bm,
		SendAt: m.Timestamp,
	}, nil
}

type BetsWinReward struct {
	Message
	Reward displayMessage `json:"bets_win_reward"`
}

func NewBetsWinProto(seq int64, rid []int32, user User, gameName string) (*logicpb.PushMsg, error) {
	now := time.Now()
	m := Message{
		Id:        seq,
		Type:      MESSAGE_TYPE,
		User:      user,
		Display:   displayByBetsWin(user, gameName),
		Time:      now.Format("15:04:05"),
		Timestamp: now.Unix(),
	}
	return m.ToRoomProto(rid)
}

func NewBetsWinRewardProto(seq int64, keys []string, user User, amount float64, buttonName string) (*logicpb.PushMsg, error) {
	now := time.Now()
	msg := "恭喜您中奖 金额＄" + money.FormatFloat64(amount) + " "
	m := BetsWinReward{
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

	bm, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	return &logicpb.PushMsg{
		Seq:    seq,
		Type:   logicpb.PushMsg_PUSH,
		Op:     pb.OpRaw,
		Keys:   keys,
		Msg:    bm,
		SendAt: m.Timestamp,
	}, nil
}
