package scheme

import (
	"strconv"
	"unicode/utf8"

	"gitlab.com/jetfueltw/cpw/alakazam/models"

	"gitlab.com/jetfueltw/cpw/alakazam/member"
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

	// Display Title 背景色
	DisplayTitleBackgroundColor = "#F85656"

	// Display 會員等級 背景色
	DisplayLevelBackgroundColor = "#7FC355"
)

// 用戶Display
func displayByUser(user User, message string) Display {
	return Display{
		User: displayUser{
			Text:   user.Name,
			Color:  USER_COLOR,
			Avatar: user.Avatar,
		},
		//TODO 會員等級
		Title: displayText{
			Text:            member.GeneralMember,
			Color:           MESSAGE_COLOR,
			BackgroundColor: DisplayLevelBackgroundColor,
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
			Text:            member.Private,
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

// 私密回應Display
func displayByPrivateReply(user User) Display {
	return Display{
		User: displayUser{
			Text:   user.Name,
			Color:  MESSAGE_SYSTEM_COLOR,
			Avatar: user.Avatar,
		},
		Title: displayText{
			Text:            member.Private,
			Color:           MESSAGE_COLOR,
			BackgroundColor: "#F79EB6",
		},
		Message: displayMessage{
			Text:            "对 " + user.Name + " 送出私密讯息",
			Color:           MESSAGE_SYSTEM_COLOR,
			BackgroundColor: NONE_COLOR,
			Entity: []textEntity{
				usernameTextEntity(user.Name, 1),
			},
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
		//TODO 會員等級
		Title: displayText{
			Text:            member.GeneralMember,
			Color:           MESSAGE_COLOR,
			BackgroundColor: DisplayLevelBackgroundColor, //"#38A2DB",
		},
		IsManage: true,
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
		//TODO 會員等級
		Title: displayText{
			Text:            member.Anchor,
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
			Text:            member.Win,
			Color:           MESSAGE_COLOR,
			BackgroundColor: DisplayTitleBackgroundColor,
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

// 等級提升Display
func displayByLevelUp(user *models.Member, level int) Display {
	return Display{
		Title: displayText{
			Text:            member.System,
			Color:           MESSAGE_COLOR,
			BackgroundColor: DisplayTitleBackgroundColor,
		},
		Message: displayMessage{
			Text:            "恭喜等級提升到" + strconv.Itoa(level),
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
			Text:            user.Name + "打赏主播" + strconv.FormatFloat(amount, 'f', -1, 64) + "钻",
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
func DisplayBySetManage(username string, set bool) Display {
	var msg string
	if set {
		msg = "用户" + username + "已被主播设置为"
	} else {
		msg = "用户" + username + "已被主播解除"
	}

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

// 禁言通知Display
func DisplayBySetBanned(username string, expired int, set bool) Display {
	var text string
	msg := "用户" + username + "已被"
	if set {
		text += msg + "禁言" + strconv.Itoa(expired/60) + "分钟"
	} else {
		msg += "解除"
		text = msg + "禁言"
	}

	return Display{
		Title: displayText{
			Text:            member.System,
			Color:           MESSAGE_COLOR,
			BackgroundColor: SYSTEM_BACKGROUND_COLOR,
		},
		Message: displayMessage{
			Text:            text,
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

// 主播解封Display
func DisplayByUnBlock(username string, expired int, set bool) Display {
	var text string
	msg := "用户" + username + "已被"
	if set {
		text += msg + "封鎖" // + strconv.Itoa(expired/60) + "分钟"
	} else {
		msg += "解除"
		text = msg + "封鎖"
	}

	return Display{
		Title: displayText{
			Text:            member.System,
			Color:           MESSAGE_COLOR,
			BackgroundColor: SYSTEM_BACKGROUND_COLOR,
		},
		Message: displayMessage{
			Text:            text,
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
func displayByConnect(level int32, isManage bool, username string) Display {
	return Display{
		Level: displayText{
			Text:            strconv.Itoa(int(level)), // TODO 會員等級
			Color:           MESSAGE_COLOR,
			BackgroundColor: DisplayLevelBackgroundColor,
		},
		IsManage: isManage,
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
		Color:           MESSAGE_SYSTEM_COLOR,
		BackgroundColor: DisplayTitleBackgroundColor,
	}
}
