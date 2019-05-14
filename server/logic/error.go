package logic

import "errors"

var (
	ConnectError = errors.New("进入聊天室失败")

	BannedError = errors.New("您在禁言状态，无法进入聊天室")
)
