package logic

import (
	"github.com/stretchr/testify/assert"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"testing"
)

func TestPushRoomNotLogin(t *testing.T) {
	p := &PushRoomForm{
		Uid:     "",
		Key:     "",
		Message: "",
	}
	err := l.PushRoom(p)
	assert.Equal(t, errors.LoginError, err)
}

func TestPushRoomNotRoomIdEmpty(t *testing.T) {
	addUser(t, "123", "", "", "test", 0)

	p := &PushRoomForm{
		Uid:     "123",
		Key:     "",
		Message: "",
	}
	err := l.PushRoom(p)
	assert.Equal(t, errors.RoomError, err)
}

func TestPushRoomNotKey(t *testing.T) {
	addUser(t, "123", "1", "1000", "test", 0)

	p := &PushRoomForm{
		Uid:     "123",
		Key:     "",
		Message: "",
	}
	err := l.PushRoom(p)
	assert.Equal(t, errors.RoomError, err)
}

func TestPushRoomNotUid(t *testing.T) {
	addUser(t, "123", "1", "1000", "test", 0)

	p := &PushRoomForm{
		Uid:     "",
		Key:     "1",
		Message: "",
	}
	err := l.PushRoom(p)
	assert.Equal(t, errors.LoginError, err)
}
