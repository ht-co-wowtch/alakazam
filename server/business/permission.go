package business

const (
	// 封鎖
	Blockade = 0

	// 查看聊天
	look = 1

	// 聊天
	message = 2

	// 發紅包
	sendBonus = 4

	// 搶紅包
	getBonus = 8

	// 發跟注
	sendFollow = 16

	// 跟注
	getFollow = 32

	// 充值
	recharge = 64

	// 打碼量
	dml = 128

	// 訊息頂置
	messageTop = 256

	// 一般權限
	PlayDefaultPermission = look + message + sendBonus + getBonus + sendFollow + getFollow + recharge + dml

	// 試玩權限
	GuestDefaultPermission = look

	// 營運權限
	marketDefaultPermission = look + message + sendBonus + getBonus + sendFollow + getFollow

	// 機器人權限
	fakeDefaultPermission = sendFollow

	// 後台權限
	adminDefaultPermission = message + sendBonus + messageTop
)

// 是否禁言
func IsBanned(weight int) bool {
	return (message & weight) != message
}

// 是否可查看聊天
func IsLook(weight int) bool {
	return (look & weight) == look
}

// 是否可以發紅包
func IsSendBonus(weight int) bool {
	return (sendBonus & weight) == sendBonus
}

// 是否可以搶紅包
func IsGetBonus(weight int) bool {
	return (getBonus & weight) == getBonus
}

// 是否可以發跟注
func IsSendFollow(weight int) bool {
	return (sendFollow & weight) == sendFollow
}

// 是否可以跟注
func IsGetFollow(weight int) bool {
	return (getFollow & weight) == getFollow
}

// 用戶權限
type Permission struct {
	Message    bool `json:"message"`
	SendBonus  bool `json:"send_bonus"`
	GetBonus   bool `json:"get_bonus"`
	SendFollow bool `json:"send_follow"`
	GetFollow  bool `json:"get_follow"`
}

// 建立用戶權限結構
func NewPermission(weight int) *Permission {
	return &Permission{
		Message:    !IsBanned(weight),
		SendBonus:  IsSendBonus(weight),
		GetBonus:   IsGetBonus(weight),
		SendFollow: IsSendFollow(weight),
		GetFollow:  IsGetFollow(weight),
	}
}
