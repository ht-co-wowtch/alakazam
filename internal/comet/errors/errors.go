package errors

import (
	"errors"
)

var (
	// bucket
	ErrBroadCastArg     = errors.New("rpc broadcast arg error")
	ErrBroadCastRoomArg = errors.New("rpc broadcast  room arg error")

	// room
	ErrRoomDroped = errors.New("room droped")
)
