package scheme

import (
	"encoding/json"
	logicpb "gitlab.com/jetfueltw/cpw/alakazam/app/logic/pb"
	"gitlab.com/jetfueltw/cpw/alakazam/member"
	"time"
)

const (
	// 普通訊息
	MESSAGE_TYPE = "message"

	// 紅包訊息
	RED_ENVELOPE_TYPE = "red_envelope"

	// 公告訊息
	TOP_TYPE = "top"

	// 禮物/打賞
	GIFT_TYPE = "gift"

	// 投注中獎打賞
	BETS_WIN_REWARD = "bets_win_reward"
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

func (m Message) ToProto() (*logicpb.PushMsg, error) {
	bm, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	return &logicpb.PushMsg{
		Seq:     m.Id,
		Msg:     bm,
		Message: m.Message,
		SendAt:  m.Timestamp,
	}, nil
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

	// 背景色
	BackgroundColor interface{} `json:"background_color"`

	// 背景圖像
	BackgroundImage interface{} `json:"background_image"`
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

	// 字體背景
	BackgroundColor string `json:"background_color"`

	// 文字實體
	Entity []TextEntity `json:"entity"`
}

// 文字實體
type TextEntity struct {
	// 類型
	Type string `json:"type"`

	// 第幾個Offset
	Offset int `json:"offset"`

	// 從Offset算幾個Length
	Length int `json:"length"`

	// 文字顏色
	Color string `json:"color"`

	// 該範圍背景顏色
	BackgroundColor string `json:"background_color"`
}

type NullDisplayMessage DisplayMessage

func (d NullDisplayMessage) MarshalJSON() ([]byte, error) {
	if d.Text == "" {
		return []byte(`null`), nil
	}
	return json.Marshal(DisplayMessage(d))
}

// 漸層色背景
type BackgroundImage struct {
	Type string `json:"type"`

	// 漸層方向
	To string `json:"to"`

	// 顏色組合
	Color map[int]string `json:"color"`
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

func (u User) ToUser(seq int64, message string) Message {
	b := u.toBase(seq, message)
	b.Type = MESSAGE_TYPE
	b.Display = displayByUser(u, message)
	return b
}

func (u User) ToStreamer(seq int64, message string) Message {
	b := u.toBase(seq, message)
	b.Type = MESSAGE_TYPE
	b.Display = displayByStreamer(u, message)
	return b
}

func (u User) ToAdmin(seq int64, message string) Message {
	b := u.toBase(seq, message)
	b.Type = MESSAGE_TYPE
	b.Display = displayByAdmin(message)
	return b
}

func (u User) ToTop(seq int64, message string) Message {
	msg := u.ToSystem(seq, message)
	msg.Type = TOP_TYPE
	return msg
}

func (u User) ToSystem(seq int64, message string) Message {
	b := u.toBase(seq, message)
	b.Type = MESSAGE_TYPE
	b.Display = displayBySystem(message)
	return b
}

func (u User) toBase(seq int64, message string) Message {
	now := time.Now()
	return Message{
		Id:        seq,
		User:      NullUser(u),
		Time:      now.Format("15:04:05"),
		Timestamp: now.Unix(),

		Uid:     u.Uid,
		Name:    u.Name,
		Avatar:  u.Avatar,
		Message: message,
	}
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
		Avatar: "root",
	}
}

func ToMessage(msgByte []byte) (Message, error) {
	var msg Message
	err := json.Unmarshal(msgByte, &msg)
	return msg, err
}

func NewConnect(seq int64, username string) Message {
	now := time.Now()
	return Message{
		Id:        seq,
		Type:      "hint",
		Display:   displayByConnect(username),
		Time:      now.Format("15:04:05"),
		Timestamp: now.Unix(),
	}
}
