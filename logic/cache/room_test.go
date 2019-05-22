package cache

import (
	"github.com/gomodule/redigo/redis"
	"github.com/rafaeljusto/redigomock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetRoom(t *testing.T) {
	roomId := "a1b4bbec3f624ecf84858632a730c688"

	mockGetRoom(roomId, []byte(`1`))

	i, err := d.GetRoom(roomId)

	assert.Nil(t, err)
	assert.Equal(t, 1, i)
}

func TestGetRoomEmpty(t *testing.T) {
	roomId := "a1b4bbec3f624ecf84858632a730c688"

	mockGetRoom(roomId, nil)

	i, err := d.GetRoom(roomId)

	assert.Equal(t, redis.ErrNil, err)
	assert.Equal(t, 0, i)
}

func TestSetRoom(t *testing.T) {
	roomId := "a1b4bbec3f624ecf84858632a730c688"

	mockSetRoom(roomId)

	err := d.SetRoom(roomId, 1)

	assert.Nil(t, err)
}

func mockSetRoom(roomId string) {
	mock.Command("SET", keyRoom(roomId), 1).
		Expect("")
	mock.Command("EXPIRE", keyRoom(roomId), 60*60).
		Expect("")
}

func mockGetRoom(roomId string, expect interface{}) *redigomock.Cmd {
	return mock.Command("GET", keyRoom(roomId)).
		Expect(expect)
}
