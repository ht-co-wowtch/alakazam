package test

import (
	"github.com/stretchr/testify/assert"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/store"
	"gitlab.com/jetfueltw/cpw/alakazam/test/internal/request"
	"testing"
)

func TestSetRoomIsBanned(t *testing.T) {
	room := store.Room{
		RoomId:    1000,
		IsMessage: false,
	}

	request.SetRoom(room)

	a, _ := request.DialAuth("1000")

	r := request.PushRoom(a.Uid, a.Key, "測試")
	e := request.ToError(t, r.Body)

	assert.Equal(t, errors.RoomBannedError.Code, e.Code)
	assert.Equal(t, errors.RoomBannedError.Message, e.Message)
}
