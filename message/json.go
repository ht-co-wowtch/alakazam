package message

import (
	"encoding/json"
	"gitlab.com/jetfueltw/cpw/alakazam/member"
)

type System struct {
	Id        int64                  `json:"id"`
	Type      string                 `json:"type"`
	Name      string                 `json:"name"`
	Message   string                 `json:"message"`
	Time      string                 `json:"time"`
	Timestamp int64                  `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

// 訊息格式
type Message struct {
	// 訊息id
	Id int64 `json:"id"`

	// 訊息種類
	Type string `json:"type"`

	// 訊息產生時間
	Time string `json:"time"`

	// 訊息產生時間戳記
	Timestamp int64 `json:"timestamp"`

	// 顯示訊息的資料
	Display Display `json:"display"`

	// 發訊息者
	User NullUser `json:"user"`

	// TODO 以下待廢棄
	Uid     string `json:"uid"`
	Name    string `json:"name"`
	Avatar  string `json:"avatar"`
	Message string `json:"message"`
}

func (m Message) MarshalBinary() (data []byte, err error) {
	return json.Marshal(m)
}

func (m Message) Score() float64 {
	return float64(m.Timestamp)
}

// 顯示訊息資料格式
type Display struct {
	// 顯示用戶
	User NullDisplayUser `json:"user"`

	// 顯示用戶等級
	Level NullDisplayText `json:"level"`

	// 顯示用戶標題
	Title NullDisplayText `json:"title"`

	// 顯示訊息內容
	Message NullDisplayMessage `json:"message"`
}

// 顯示用戶資料
type DisplayUser struct {
	// 用戶名
	Text string `json:"text"`

	// 字體顏色
	Color string `json:"color"`

	// 用戶頭像
	Avatar string `json:"avatar"`
}

type NullDisplayUser DisplayUser

func (d NullDisplayUser) MarshalJSON() ([]byte, error) {
	if d.Text == "" {
		return []byte(`null`), nil
	}
	return json.Marshal(DisplayUser(d))
}

// 文字資料
type DisplayText struct {
	// 文字
	Text string `json:"text"`

	// 字體顏色(預設)
	Color string `json:"color"`

	// 字體背景
	BackgroundColor string `json:"background_color"`
}

type NullDisplayText DisplayText

func (d NullDisplayText) MarshalJSON() ([]byte, error) {
	if d.Text == "" {
		return []byte(`null`), nil
	}
	return json.Marshal(DisplayText(d))
}

type DisplayMessage struct {
	// 文字
	Text string `json:"text"`

	// 字體顏色(預設)
	Color string `json:"color"`

	// 各範圍文字的顏色
	PartColor []PartColor `json:"part_color"`
}

type PartColor struct {
	Offset int    `json:"offset"`
	Length int    `json:"length"`
	Value  string `json:"value"`
}

type NullDisplayMessage DisplayMessage

func (d NullDisplayMessage) MarshalJSON() ([]byte, error) {
	if d.Text == "" {
		return []byte(`null`), nil
	}
	return json.Marshal(DisplayMessage(d))
}

// 用戶資料
type User struct {
	Id int64 `json:"-"`

	// uid
	Uid string `json:"uid"`

	// 名稱
	Name string `json:"name"`

	// 頭像
	Avatar string `json:"avatar"`
}

type NullUser User

func (d NullUser) MarshalJSON() ([]byte, error) {
	if d.Uid == "" || d.Name == "" {
		return []byte(`null`), nil
	}
	return json.Marshal(User(d))
}

// 管理員用戶資料
func NewRoot() User {
	return User{
		Id:     member.RootMid,
		Uid:    member.RootUid,
		Name:   member.RootName,
		Avatar: avatarRoot,
	}
}

// 紅包訊息格式
type RedEnvelopeMessage struct {
	// 訊息格式
	Message

	// 紅包資料
	RedEnvelope RedEnvelope `json:"red_envelope"`
}

func (m RedEnvelopeMessage) Score() float64 {
	return float64(m.Message.Timestamp)
}

// 紅包資料
type RedEnvelope struct {
	// 紅包id
	Id string `json:"id"`

	// 紅包token
	Token string `json:"token"`

	// 紅包多久過期
	Expired string `json:"expired"`
}

func (m RedEnvelope) MarshalBinary() (data []byte, err error) {
	return json.Marshal(m)
}

// 跟投訊息格式
type Bets struct {
	// 訊息id
	Id int64 `json:"id"`

	// 訊息種類
	Type string `json:"type"`

	// 訊息產生時間
	Time string `json:"time"`

	// 訊息產生時間戳記
	Timestamp int64 `json:"timestamp"`

	// 顯示訊息的資料
	Display Display `json:"display"`

	// 發訊息者
	User User `json:"user"`

	// 跟投資料
	Bet Bet `json:"bet"`

	// TODO 以下待廢棄
	Uid          string     `json:"uid"`
	Name         string     `json:"name"`
	Avatar       string     `json:"avatar"`
	GameId       int        `json:"game_id"`
	PeriodNumber int        `json:"period_number"`
	Items        []BetOrder `json:"bets"`
	Count        int        `json:"count"`
	TotalAmount  int        `json:"total_amount"`
}

// 跟投資料
type Bet struct {
	GameId       int        `json:"game_id"`
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
