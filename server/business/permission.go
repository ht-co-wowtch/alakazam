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
	bonus = 8

	// 發跟注
	sendFollow = 16

	// 跟注
	follow = 32

	// 充值
	recharge = 64

	// 打碼量
	dml = 128

	// 訊息頂置
	messageTop = 256

	// 一般權限
	PlayDefaultPermission = look + message + sendBonus + bonus + sendFollow + follow + recharge + dml

	// 試玩權限
	GuestDefaultPermission = look

	// 營運權限
	marketDefaultPermission = look + message + sendBonus + bonus + sendFollow + follow

	// 機器人權限
	fakeDefaultPermission = sendFollow

	// 後台權限
	adminDefaultPermission = message + sendBonus + messageTop
)

// 是否禁言
func IsBanned(weight int) bool {
	return (message & weight) != message
}
