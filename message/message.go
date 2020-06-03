package message

import (
	"database/sql"
	"encoding/json"
	"github.com/go-redis/redis"
	"gitlab.com/jetfueltw/cpw/alakazam/app/logic/pb"
	"gitlab.com/jetfueltw/cpw/alakazam/member"
	"gitlab.com/jetfueltw/cpw/alakazam/models"
	"strconv"
	"time"
)

type History struct {
	db     *models.Store
	member *member.Member
	cache  Cache
}

func NewHistory(db *models.Store, c *redis.Client, member *member.Member) *History {
	return &History{
		db:     db,
		cache:  newCache(c),
		member: member,
	}
}

const (
	// 普通訊息
	MessageType = "message"

	// 紅包訊息
	RedEnvelopeType = "red_envelope"

	// 公告訊息
	TopType = "top"

	// 跟注
	BetsType = "bets"

	// 禮物
	GiftType = "gift"
)

func (h *History) Get(roomId int32, at time.Time) ([]interface{}, error) {
	if time.Now().Add(-2 * time.Hour).After(at) {
		return []interface{}{}, nil
	}

	msgs, err := h.cache.getMessage(roomId, at)
	if err != nil {
		return []interface{}{}, nil
	}

	if len(msgs) > 0 {
		message := make([]interface{}, 0, len(msgs))
		for i := len(msgs); i > 0; i-- {
			b := msgs[i-1]
			message = append(message, stringJson(b))
		}
		return message, nil
	}

	msg, err := h.db.GetRoomMessage(roomId, at)
	if err != nil {
		if err != sql.ErrNoRows {
			return nil, err
		}
		return []interface{}{}, nil
	}

	mids := make([]int, len(msg.Message)+len(msg.RedEnvelopeMessage))

	for _, v := range msg.Message {
		mids = append(mids, v.MemberId)
	}
	for _, v := range msg.RedEnvelopeMessage {
		mids = append(mids, v.MemberId)
	}

	ms, err := h.member.GetMembers(mids)
	if err != nil {
		return []interface{}{}, err
	}

	memberMap := make(map[int]models.Member, len(ms))
	for _, v := range ms {
		memberMap[v.Id] = v
	}

	data := make([]interface{}, 0)
	for _, msgId := range msg.List {
		switch msg.Type[msgId] {
		case pb.PushMsg_MONEY:
			redEnvelope := msg.RedEnvelopeMessage[msgId]
			user := memberMap[redEnvelope.MemberId]
			data = append(data, RedEnvelopeMessage{
				Message: Message{
					Id:        msgId,
					Type:      RedEnvelopeType,
					Time:      redEnvelope.SendAt.Format("15:04:05"),
					Timestamp: redEnvelope.SendAt.Unix(),
					Display: Display{
						Message: NullDisplayMessage{
							Text:  redEnvelope.Message,
							Color: "#FFFFFF",
						},
					},
					User: NullUser{
						Uid:    user.Uid,
						Name:   user.Name,
						Avatar: ToAvatarName(user.Gender),
					},

					Uid:     user.Uid,
					Name:    user.Name,
					Avatar:  ToAvatarName(user.Gender),
					Message: redEnvelope.Message,
				},
				RedEnvelope: RedEnvelope{
					Id:      redEnvelope.RedEnvelopesId,
					Token:   redEnvelope.Token,
					Expired: redEnvelope.ExpireAt.Format(time.RFC3339),
				},
			})
		case pb.PushMsg_USER:
			user := memberMap[msg.Message[msgId].MemberId]
			data = append(data, Message{
				Id:        msgId,
				Uid:       user.Uid,
				Name:      user.Name,
				Type:      MessageType,
				Avatar:    ToAvatarName(user.Gender),
				Message:   msg.Message[msgId].Message,
				Time:      msg.Message[msgId].SendAt.Format("15:04:05"),
				Timestamp: msg.Message[msgId].SendAt.Unix(),
			})
		case pb.PushMsg_ADMIN:
			user := memberMap[msg.Message[msgId].MemberId]
			data = append(data, Message{
				Id:        msgId,
				Uid:       user.Uid,
				Name:      user.Name,
				Type:      MessageType,
				Avatar:    avatarRoot,
				Message:   msg.Message[msgId].Message,
				Time:      msg.Message[msgId].SendAt.Format("15:04:05"),
				Timestamp: msg.Message[msgId].SendAt.Unix(),
			})
		}
	}

	if len(data) == 0 {
		return data, nil
	}
	return data, h.cache.addMessages(roomId, data)
}

type stringJson string

func (s stringJson) MarshalJSON() ([]byte, error) {
	return []byte(s), nil
}

func RoomTopMessageToMessage(msg models.RoomTopMessage) Message {
	return Message{
		Id:        msg.MsgId,
		Uid:       member.RootUid,
		Type:      TopType,
		Message:   msg.Message,
		Time:      msg.SendAt.Format("15:04:05"),
		Timestamp: msg.SendAt.Unix(),
		Display:   DisplayByMessage(msg.Message),
	}
}

func RoomBulletinMessageToMessage(msg models.RoomTopMessage) Message {
	return Message{
		Id:        msg.MsgId,
		Uid:       member.RootUid,
		Type:      MessageType,
		Message:   msg.Message,
		Time:      msg.SendAt.Format("15:04:05"),
		Timestamp: msg.SendAt.Unix(),
		Display:   DisplayBySystem(msg.Message),
	}
}

func ToMessage(msgByte []byte) (Message, error) {
	var msg Message
	err := json.Unmarshal(msgByte, &msg)
	return msg, err
}

const (
	NONE_COLOR = "#00000000"

	// 會員名稱文字顏色
	USER_COLOR = "#7CE7EB"

	// 訊息文字顏色
	MESSAGE_COLOR = "#FFFFFF"

	// 訊息內用戶名字體顏色
	MESSAGE_USERNAME_COLOR = "#7CE7EB"

	// 訊息框 背景色
	MESSAGE_BACKGROUND_COLOR = "#0000003f"

	// 系統訊息字體顏色
	MESSAGE_SYSTEM_COLOR = "#FFFFAA"

	// 系統框 背景色
	SYSTEM_BACKGROUND_COLOR = "#FC8813"
)

// 用戶Display
func DisplayByUser(user User, message string) Display {
	return Display{
		User: NullDisplayUser{
			Text:   user.Name,
			Color:  USER_COLOR,
			Avatar: user.Avatar,
		},
		Level: NullDisplayText{
			Text:            "会员",
			Color:           MESSAGE_COLOR,
			BackgroundColor: "#7FC355",
		},
		Message: NullDisplayMessage{
			Text:            message,
			Color:           MESSAGE_COLOR,
			BackgroundColor: NONE_COLOR,
		},
		BackgroundColor: MESSAGE_BACKGROUND_COLOR,
	}
}

// 主播Display
func DisplayByStreamer(user User, message string) Display {
	return Display{
		User: NullDisplayUser{
			Text:   user.Name,
			Color:  MESSAGE_SYSTEM_COLOR,
			Avatar: user.Avatar,
		},
		Title: NullDisplayText{
			Text:            "主播",
			Color:           MESSAGE_COLOR,
			BackgroundColor: "#B57AA8",
		},
		Message: NullDisplayMessage{
			Text:            message,
			Color:           MESSAGE_COLOR,
			BackgroundColor: NONE_COLOR,
		},
		BackgroundColor: "#B57AA87F",
	}
}

// 管理員Display
func DisplayByAdmin(user User, message string) Display {
	return Display{
		User: NullDisplayUser{
			Text:   user.Name,
			Color:  USER_COLOR,
			Avatar: user.Avatar,
		},
		Title: NullDisplayText{
			Text:            member.RootName,
			Color:           MESSAGE_COLOR,
			BackgroundColor: "#7FC355",
		},
		Message: NullDisplayMessage{
			Text:            message,
			Color:           MESSAGE_COLOR,
			BackgroundColor: NONE_COLOR,
		},
		BackgroundColor: MESSAGE_BACKGROUND_COLOR,
	}
}

// 跟投Display
func DisplayByBets(user User, gameName string, amount int) Display {
	return Display{
		User: NullDisplayUser{
			Text:   user.Name,
			Color:  USER_COLOR,
			Avatar: user.Avatar,
		},
		Title: NullDisplayText{
			Text:            member.System,
			Color:           MESSAGE_COLOR,
			BackgroundColor: SYSTEM_BACKGROUND_COLOR,
		},
		Message: NullDisplayMessage{
			Text:            "用戶" + user.Name + "在" + gameName + "下注" + string(amount) + "元",
			Color:           MESSAGE_SYSTEM_COLOR,
			BackgroundColor: NONE_COLOR,
			Entity: []Entity{
				buttonEntity(user.Name, 2),
			},
		},
		BackgroundColor: MESSAGE_BACKGROUND_COLOR,
	}
}

// 注單派彩Display
func DisplayByBetsPay(user User, gameName string) Display {
	return Display{
		User: NullDisplayUser{
			Text:   user.Name,
			Color:  USER_COLOR,
			Avatar: user.Avatar,
		},
		Title: NullDisplayText{
			Text:            member.System,
			Color:           MESSAGE_COLOR,
			BackgroundColor: SYSTEM_BACKGROUND_COLOR,
		},
		Message: NullDisplayMessage{
			Text:            "用戶" + user.Name + "在" + gameName + "赢得奖了",
			Color:           MESSAGE_SYSTEM_COLOR,
			BackgroundColor: NONE_COLOR,
			Entity: []Entity{
				usernameEntity(user.Name, 2),
			},
		},
		BackgroundColor: MESSAGE_BACKGROUND_COLOR,
	}
}

// 禮物 Display
func DisplayByGift(user User, name string) Display {
	return Display{
		Title: NullDisplayText{
			Text:            member.System,
			Color:           MESSAGE_COLOR,
			BackgroundColor: SYSTEM_BACKGROUND_COLOR,
		},
		Message: NullDisplayMessage{
			Text:            user.Name + "送出" + name + "x1",
			Color:           MESSAGE_SYSTEM_COLOR,
			BackgroundColor: NONE_COLOR,
			Entity: []Entity{
				usernameEntity(user.Name, 0),
			},
		},
		BackgroundColor: MESSAGE_BACKGROUND_COLOR,
	}
}

// 打賞 Display
func DisplayByReward(user User, amount float64) Display {
	return Display{
		Title: NullDisplayText{
			Text:            member.System,
			Color:           MESSAGE_COLOR,
			BackgroundColor: SYSTEM_BACKGROUND_COLOR,
		},
		Message: NullDisplayMessage{
			Text:            user.Name + "打賞主播" + strconv.FormatFloat(amount, 'f', -1, 64) + "元",
			Color:           MESSAGE_SYSTEM_COLOR,
			BackgroundColor: NONE_COLOR,
			Entity: []Entity{
				usernameEntity(user.Name, 0),
			},
		},
		BackgroundColor: MESSAGE_BACKGROUND_COLOR,
	}
}

// 系統Display
func DisplayBySystem(message string) Display {
	return Display{
		Title: NullDisplayText{
			Text:            member.System,
			Color:           MESSAGE_COLOR,
			BackgroundColor: SYSTEM_BACKGROUND_COLOR,
		},
		Message: NullDisplayMessage{
			Text:            message,
			Color:           MESSAGE_SYSTEM_COLOR,
			BackgroundColor: NONE_COLOR,
		},
		BackgroundColor: MESSAGE_BACKGROUND_COLOR,
	}
}

// 單純訊息Display
func DisplayByMessage(message string) Display {
	return Display{
		Message: NullDisplayMessage{
			Text:            message,
			Color:           MESSAGE_COLOR,
			BackgroundColor: NONE_COLOR,
		},
		BackgroundColor: MESSAGE_BACKGROUND_COLOR,
	}
}

func usernameEntity(name string, offset int) Entity {
	return Entity{
		Type:            "username",
		Offset:          offset,
		Length:          len(name),
		Color:           MESSAGE_USERNAME_COLOR,
		BackgroundColor: NONE_COLOR,
	}
}

func buttonEntity(name string, offset int) Entity {
	return Entity{
		Type:            "button",
		Offset:          offset,
		Length:          len(name),
		Color:           MESSAGE_COLOR,
		BackgroundColor: "#F85656",
	}
}
