package logic

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPushKeys(t *testing.T) {
	var (
		c    = context.TODO()
		keys = []string{"test_key"}
		msg  = []byte("hello")
	)
	err := lg.PushKeys(c, keys, msg)
	assert.Nil(t, err)
}

func TestPushRoom(t *testing.T) {
	var (
		c    = context.TODO()
		room = "test_room"
		msg  = []byte("hello")
	)
	err := lg.PushRoom(c, room, msg)
	assert.Nil(t, err)
}

func TestPushAll(t *testing.T) {
	var (
		c     = context.TODO()
		speed = int32(100)
		msg   = []byte("hello")
	)
	err := lg.PushAll(c, speed, msg)
	assert.Nil(t, err)
}
