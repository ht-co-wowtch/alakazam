package scheme

import (
	"gitlab.com/jetfueltw/cpw/alakazam/member"
	"strconv"
)

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
func displayByUser(user User, message string) Display {
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
func displayByStreamer(user User, message string) Display {
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
func displayByAdmin(user User, message string) Display {
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
func displayByBets(user User, gameName string, amount int) Display {
	msg := "用戶" + user.Name + "在" + gameName + "下注" + strconv.Itoa(amount) + "元"
	return Display{
		Title: NullDisplayText{
			Text:            member.System,
			Color:           MESSAGE_COLOR,
			BackgroundColor: SYSTEM_BACKGROUND_COLOR,
		},
		Message: NullDisplayMessage{
			Text:            msg + " ＋跟注",
			Color:           MESSAGE_SYSTEM_COLOR,
			BackgroundColor: NONE_COLOR,
			Entity: []TextEntity{
				TextEntity{
					Type:            "button",
					Offset:          len(msg),
					Length:          len(" ＋跟注"),
					Color:           MESSAGE_COLOR,
					BackgroundColor: "#F85656",
				},
			},
		},
		BackgroundColor: MESSAGE_BACKGROUND_COLOR,
	}
}

// 注單派彩Display
func displayByBetsPay(user User, gameName string) Display {
	return Display{
		User: NullDisplayUser{
			Text:   user.Name,
			Color:  USER_COLOR,
			Avatar: user.Avatar,
		},
		Title: NullDisplayText{
			Text:            "中奖",
			Color:           MESSAGE_COLOR,
			BackgroundColor: SYSTEM_BACKGROUND_COLOR,
		},
		Message: NullDisplayMessage{
			Text:            "用戶" + user.Name + "在" + gameName + "赢得奖了",
			Color:           MESSAGE_SYSTEM_COLOR,
			BackgroundColor: NONE_COLOR,
			Entity: []TextEntity{
				usernameEntity(user.Name, 2),
			},
		},
		BackgroundColor: MESSAGE_BACKGROUND_COLOR,
	}
}

// 禮物 Display
func displayByGift(user User, name string) Display {
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
			Entity: []TextEntity{
				usernameEntity(user.Name, 0),
			},
		},
		BackgroundColor: MESSAGE_BACKGROUND_COLOR,
	}
}

// 打賞 Display
func displayByReward(user User, amount float64) Display {
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
			Entity: []TextEntity{
				usernameEntity(user.Name, 0),
			},
		},
		BackgroundColor: MESSAGE_BACKGROUND_COLOR,
	}
}

// 系統Display
func displayBySystem(message string) Display {
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

// 進場Display
func displayByConnect(username string) Display {
	return Display{
		Level: NullDisplayText{
			Text:            "会员",
			Color:           MESSAGE_COLOR,
			BackgroundColor: "#7FC355",
		},
		Message: NullDisplayMessage{
			Text:            username + "进入聊天室",
			Color:           MESSAGE_SYSTEM_COLOR,
			BackgroundColor: NONE_COLOR,
			Entity: []TextEntity{
				usernameEntity(username, 0),
			},
		},
		BackgroundColor: MESSAGE_BACKGROUND_COLOR,
		BackgroundImage: []interface{}{
			LinearGradientBackground{
				Type:  "linear-gradient",
				To:    "right",
				Color: []string{"#FC881380", "#FC8813"},
			},
		},
	}
}

func usernameEntity(name string, offset int) TextEntity {
	return TextEntity{
		Type:            "username",
		Offset:          offset,
		Length:          len(name),
		Color:           MESSAGE_USERNAME_COLOR,
		BackgroundColor: NONE_COLOR,
	}
}