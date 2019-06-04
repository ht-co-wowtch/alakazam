package permission

const (
	// 封鎖
	Blockade = 0

	// 查看聊天
	look = 1

	// 聊天
	Message = 2

	// 搶紅包
	getBonus = 4

	// 發跟注
	sendFollow = 8

	// 跟注
	getFollow = 16

	// 充值&打碼量
	money = 32

	// 訊息頂置
	messageTop = 256

	// 一般權限
	PlayDefaultPermission = look + Message + getBonus + sendFollow + getFollow

	// 試玩權限
	GuestDefaultPermission = look

	// 營運權限
	marketDefaultPermission = look + Message + getBonus + sendFollow + getFollow

	// 後台權限
	adminDefaultPermission = Message + messageTop
)

// 是否禁言
func IsBanned(weight int) bool {
	return (Message & weight) != Message
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

// 是否有充值&打碼量限制
func IsMoney(weight int) bool {
	return (money & weight) == money
}

// 用戶權限
type Permission struct {
	Message    bool `json:"Message"`
	SendBonus  bool `json:"send_bonus"`
	GetBonus   bool `json:"get_bonus"`
	SendFollow bool `json:"send_follow"`
	GetFollow  bool `json:"get_follow"`
}

// 建立用戶權限結構
func NewPermission(weight int) *Permission {
	return &Permission{
		Message:    !IsBanned(weight),
		SendBonus:  true,
		GetBonus:   IsGetBonus(weight),
		SendFollow: IsSendFollow(weight),
		GetFollow:  IsGetFollow(weight),
	}
}
