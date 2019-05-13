package errors

import (
	"errors"
)

var (
	// ring
	ErrRingEmpty = errors.New("ring buffer empty")
	ErrRingFull  = errors.New("ring buffer full")

	// channel
	ErrPushMsgArg = errors.New("rpc pushmsg arg error")

	// bucket
	ErrBroadCastArg     = errors.New("rpc broadcast arg error")
	ErrBroadCastRoomArg = errors.New("rpc broadcast room arg error")

	// room
	ErrRoomDroped = errors.New("room droped")

	Blockade = "您在封鎖状态，无法进入聊天室"

	BlockadeError = errors.New(Blockade)
)
