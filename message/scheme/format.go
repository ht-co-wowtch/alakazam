package scheme

import (
	"gitlab.com/jetfueltw/cpw/alakazam/member"
	"strconv"
	"unicode/utf8"
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
		User: displayUser{
			Text:   user.Name,
			Color:  USER_COLOR,
			Avatar: user.Avatar,
		},
		Level: displayText{
			Text:            "会员",
			Color:           MESSAGE_COLOR,
			BackgroundColor: "#7FC355",
		},
		Message: displayMessage{
			Text:            message,
			Color:           MESSAGE_COLOR,
			BackgroundColor: NONE_COLOR,
		},
		BackgroundColor: MESSAGE_BACKGROUND_COLOR,
	}
}

// 私密Display
func displayByPrivate(user User, message string) Display {
	return Display{
		User: displayUser{
			Text:   user.Name,
			Color:  MESSAGE_SYSTEM_COLOR,
			Avatar: user.Avatar,
		},
		Title: displayText{
			Text:            "私讯",
			Color:           MESSAGE_COLOR,
			BackgroundColor: "#F79EB6",
		},
		Message: displayMessage{
			Text:            message,
			Color:           MESSAGE_COLOR,
			BackgroundColor: NONE_COLOR,
		},
		BackgroundColor: "#38A2DB7F",
	}
}

// 房管Display
func displayByManage(user User, message string) Display {
	return Display{
		User: displayUser{
			Text:   user.Name,
			Color:  USER_COLOR,
			Avatar: user.Avatar,
		},
		Level: displayText{
			Text:            "房管",
			Color:           MESSAGE_COLOR,
			BackgroundColor: "#38A2DB",
		},
		Message: displayMessage{
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
		User: displayUser{
			Text:   user.Name,
			Color:  MESSAGE_SYSTEM_COLOR,
			Avatar: user.Avatar,
		},
		Title: displayText{
			Text:            "主播",
			Color:           MESSAGE_COLOR,
			BackgroundColor: "#B57AA8",
		},
		Message: displayMessage{
			Text:            message,
			Color:           MESSAGE_COLOR,
			BackgroundColor: NONE_COLOR,
		},
		BackgroundColor: "#B57AA87F",
	}
}

// 管理員Display
func displayByAdmin(message string) Display {
	return Display{
		Title: displayText{
			Text:            member.RootName,
			Color:           MESSAGE_COLOR,
			BackgroundColor: "#7FC355",
		},
		Message: displayMessage{
			Text:            message,
			Color:           MESSAGE_COLOR,
			BackgroundColor: NONE_COLOR,
		},
		BackgroundColor: MESSAGE_BACKGROUND_COLOR,
	}
}

// 跟投Display
func displayByBets(user User, gameName string, amount int) Display {
	msg := "用户" + user.Name + "在" + gameName + "下注" + strconv.Itoa(amount) + "元"
	return Display{
		Title: displayText{
			Text:            member.System,
			Color:           MESSAGE_COLOR,
			BackgroundColor: SYSTEM_BACKGROUND_COLOR,
		},
		Message: displayMessage{
			Text:            msg + " ＋跟注",
			Color:           MESSAGE_SYSTEM_COLOR,
			BackgroundColor: NONE_COLOR,
			Entity: []textEntity{
				usernameTextEntity(user.Name, 2),
				buttonTextEntity(" ＋跟注", utf8.RuneCountInString(msg)),
			},
		},
		BackgroundColor: MESSAGE_BACKGROUND_COLOR,
	}
}

// 投注中獎Display
func displayByBetsWin(user User, gameName string) Display {
	return Display{
		Title: displayText{
			Text:            "中奖",
			Color:           MESSAGE_COLOR,
			BackgroundColor: "#F85656",
		},
		Message: displayMessage{
			Text:            "恭喜用户" + user.Name + "在" + gameName + "中奖了",
			Color:           MESSAGE_SYSTEM_COLOR,
			BackgroundColor: NONE_COLOR,
			Entity: []textEntity{
				usernameTextEntity(user.Name, 4),
			},
		},
		BackgroundColor: MESSAGE_BACKGROUND_COLOR,
	}
}

// 禮物 Display
func displayByGift(user User, name string) Display {
	return Display{
		Title: displayText{
			Text:            member.System,
			Color:           MESSAGE_COLOR,
			BackgroundColor: SYSTEM_BACKGROUND_COLOR,
		},
		Message: displayMessage{
			Text:            user.Name + "送出" + name + "x1",
			Color:           MESSAGE_SYSTEM_COLOR,
			BackgroundColor: NONE_COLOR,
			Entity: []textEntity{
				usernameTextEntity(user.Name, 0),
			},
		},
		BackgroundColor: MESSAGE_BACKGROUND_COLOR,
	}
}

// 打賞 Display
func displayByReward(user User, amount float64) Display {
	return Display{
		Title: displayText{
			Text:            member.System,
			Color:           MESSAGE_COLOR,
			BackgroundColor: SYSTEM_BACKGROUND_COLOR,
		},
		Message: displayMessage{
			Text:            user.Name + "打赏主播" + strconv.FormatFloat(amount, 'f', -1, 64) + "元",
			Color:           MESSAGE_SYSTEM_COLOR,
			BackgroundColor: NONE_COLOR,
			Entity: []textEntity{
				usernameTextEntity(user.Name, 0),
			},
		},
		BackgroundColor: MESSAGE_BACKGROUND_COLOR,
	}
}

// 系統Display
func displayBySystem(message string) Display {
	return Display{
		Title: displayText{
			Text:            member.System,
			Color:           MESSAGE_COLOR,
			BackgroundColor: SYSTEM_BACKGROUND_COLOR,
		},
		Message: displayMessage{
			Text:            message,
			Color:           MESSAGE_SYSTEM_COLOR,
			BackgroundColor: NONE_COLOR,
		},
		BackgroundColor: MESSAGE_BACKGROUND_COLOR,
	}
}

// 房管通知Display
func DisplayBySetManage(username string) Display {
	msg := "用户" + username + "已被主播设置为"
	return Display{
		Title: displayText{
			Text:            member.System,
			Color:           MESSAGE_COLOR,
			BackgroundColor: SYSTEM_BACKGROUND_COLOR,
		},
		Message: displayMessage{
			Text:            msg + "房管",
			Color:           MESSAGE_SYSTEM_COLOR,
			BackgroundColor: NONE_COLOR,
			Entity: []textEntity{
				usernameTextEntity(username, 2),
				textEntity{
					Type:            "text",
					Offset:          utf8.RuneCountInString(msg),
					Length:          2,
					Color:           "#FFFFFF",
					BackgroundColor: "#38A2DB",
				},
			},
		},
		BackgroundColor: MESSAGE_BACKGROUND_COLOR,
	}
}

// 進場Display
func displayByConnect(username string) Display {
	return Display{
		Level: displayText{
			Text:            "会员",
			Color:           MESSAGE_COLOR,
			BackgroundColor: "#7FC355",
		},
		Message: displayMessage{
			Text:            username + "进入聊天室",
			Color:           MESSAGE_SYSTEM_COLOR,
			BackgroundColor: NONE_COLOR,
			Entity: []textEntity{
				usernameTextEntity(username, 0),
			},
		},
		BackgroundImage: backgroundImage{
			Type: "linear-gradient",
			To:   "right",
			Color: map[int]string{
				0:  "#FC881380",
				99: "#FC881300",
			},
		},
	}
}

func usernameTextEntity(name string, offset int) textEntity {
	return textEntity{
		Type:            "username",
		Offset:          offset,
		Length:          utf8.RuneCountInString(name),
		Color:           MESSAGE_USERNAME_COLOR,
		BackgroundColor: NONE_COLOR,
	}
}

func buttonTextEntity(name string, offset int) textEntity {
	return textEntity{
		Type:            "button",
		Offset:          offset,
		Length:          utf8.RuneCountInString(name),
		Color:           "#FFFFAA",
		BackgroundColor: "#F85656",
	}
}
