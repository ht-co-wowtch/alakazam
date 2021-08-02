package scheme

import (
	"encoding/json"
	"gitlab.com/jetfueltw/cpw/alakazam/app/comet/pb"
	logicpb "gitlab.com/jetfueltw/cpw/alakazam/app/logic/pb"
	"gitlab.com/jetfueltw/cpw/alakazam/member"
	"gitlab.com/jetfueltw/cpw/alakazam/models"
	"time"
)

const (
	// 普通訊息
	MESSAGE_TYPE = "message"

	// 私密
	PRIVATE_TYPE = "private_message"

	// 紅包訊息
	RED_ENVELOPE_TYPE = "red_envelope"

	// 公告訊息
	TOP_TYPE = "top"

	// 禮物/打賞
	GIFT_TYPE = "gift"

	// 投注中獎打賞
	BETS_WIN_REWARD = "bets_win_reward"

	// 關注
	FOLLOW = "follow"
)

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
	Display interface{} `json:"display"`

	// 發訊息者
	User interface{} `json:"user"`

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
		Seq:    m.Id,
		Op:     pb.OpRaw,
		Msg:    bm,
		SendAt: m.Timestamp,
	}, nil
}

func (m Message) ToRoomProto(rid []int32) (*logicpb.PushMsg, error) {
	p, err := m.ToProto()
	if err != nil {
		return nil, err
	}
	p.Room = rid
	p.Type = logicpb.PushMsg_ROOM
	return p, nil
}

// 顯示訊息資料格式
type Display struct {
	// 顯示用戶
	User interface{} `json:"user"`

	// 顯示用戶等級
	Level interface{} `json:"level"`

	// 顯示用戶標題
	Title interface{} `json:"title"`

	// 顯示訊息內容
	Message interface{} `json:"message"`

	// 背景色
	BackgroundColor interface{} `json:"background_color"`

	// 背景圖像
	BackgroundImage interface{} `json:"background_image"`
}

// 顯示用戶資料
type displayUser struct {
	// 用戶名
	Text string `json:"text"`

	// 字體顏色
	Color string `json:"color"`

	// 用戶頭像
	Avatar string `json:"avatar"`
}

// 文字資料
type displayText struct {
	// 文字
	Text string `json:"text"`

	// 字體顏色(預設)
	Color string `json:"color"`

	// 字體背景
	BackgroundColor string `json:"background_color"`
}

type displayMessage struct {
	// 文字
	Text string `json:"text"`

	// 字體顏色(預設)
	Color string `json:"color"`

	// 字體背景
	BackgroundColor string `json:"background_color"`

	// 文字實體
	Entity []textEntity `json:"entity"`
}

// 文字實體
type textEntity struct {
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

// 漸層色背景
type backgroundImage struct {
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

	Type string `json:"type"`
}

func NewUser(member models.Member) User {
	return User{
		Id:     member.Id,
		Uid:    member.Uid,
		Name:   member.Name,
		Type:   ToType(member.Type),
		Avatar: ToAvatarName(member.Gender),
	}
}

func (u User) ToUser(seq int64, message string) Message {
	b := u.toBase(seq, message)
	b.Type = MESSAGE_TYPE
	b.Display = displayByUser(u, message)
	return b
}

// 私訊
func (u User) ToPrivate(seq int64, message string) Message {
	b := u.toBase(seq, message)
	b.Type = PRIVATE_TYPE
	b.Display = displayByPrivate(u, message)
	return b
}

// 私訊回覆
func (u User) ToPrivateReply(seq int64) Message {
	display := displayByPrivateReply(u)
	m := display.Message.(displayMessage)
	b := u.toBase(seq, m.Text)
	b.Type = PRIVATE_TYPE
	b.Display = display
	return b
}

func (u User) ToStreamer(seq int64, message string) Message {
	b := u.toBase(seq, message)
	b.Type = MESSAGE_TYPE
	b.Display = displayByStreamer(u, message)
	return b
}

// 產生一般訊息
func (u User) ToManage(seq int64, message string) Message {
	b := u.toBase(seq, message)
	b.Type = MESSAGE_TYPE
	b.Display = displayByManage(u, message)
	return b
}

func (u User) ToAdmin(seq int64, message string) Message {
	b := u.toBase(seq, message)
	b.Type = MESSAGE_TYPE
	b.Display = displayByAdmin(message)
	return b
}

// 產生置頂訊息
func (u User) ToTop(seq int64, message string) Message {
	msg := u.ToSystem(seq, message)
	msg.Type = TOP_TYPE
	return msg
}

// 產生公告訊息
func (u User) ToSystem(seq int64, message string) Message {
	b := u.toBase(seq, message)
	b.Type = MESSAGE_TYPE
	b.Display = displayBySystem(message)
	return b
}

func (u User) DisplayToMessage(seq int64, display Display) Message {
	m := display.Message.(displayMessage)
	b := u.toBase(seq, m.Text)
	b.Type = MESSAGE_TYPE
	b.Display = display
	return b
}

func (u User) toBase(seq int64, message string) Message {
	now := time.Now()
	return Message{
		Id:        seq,
		User:      u,
		Time:      now.Format("15:04:05"),
		Timestamp: now.Unix(),

		Uid:     u.Uid,
		Name:    u.Name,
		Avatar:  u.Avatar,
		Message: message,
	}
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

func NewConnect(seq int64, level, username string) Message {
	now := time.Now()
	return Message{
		Id:        seq,
		Type:      "hint",
		Display:   displayByConnect(level, username),
		Time:      now.Format("15:04:05"),
		Timestamp: now.Unix(),
	}
}

const (
	avatarFemale = "female"
	avatarMale   = "male"
	avatarOther  = "other"
	avatarRoot   = "root"
)

func ToAvatarName(code int32) string {
	switch code {
	case 0:
		return avatarFemale
	case 1:
		return avatarMale
	case 2:
		return avatarOther
	case 99:
		return avatarRoot
	}
	return avatarOther
}

func ToType(t int) string {
	switch t {
	case 1:
		return "market"
	case 2:
		return "player"
	case 3:
		return "streamer"
	case 4:
		return "manage"
	}
	return "guest"
}
