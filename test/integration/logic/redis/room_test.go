package redis

import (
	"github.com/go-redis/redis"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSetRoom(t *testing.T) {
	err := d.SetRoom("1", 1, 1, 1, 1)
	assert.Nil(t, err)
}

func TestGetRoom(t *testing.T) {
	roomId := "2"
	d.SetRoom(roomId, 255, 1, 1, 1)
	i, err := d.GetRoom(roomId)

	assert.Nil(t, err)
	assert.Equal(t, 255, i)
}

func TestGetRoomByMoney(t *testing.T) {
	day := 1
	dml := 100
	amount := 1000
	roomId := "3"
	d.SetRoom(roomId, day, day, dml, amount)
	day, dml, amount, err := d.GetRoomByMoney(roomId)

	assert.Nil(t, err)
	assert.Equal(t, day, day)
	assert.Equal(t, dml, dml)
	assert.Equal(t, amount, amount)
}

func TestGetRoomEmpty(t *testing.T) {
	i, err := d.GetRoom("4")

	assert.Equal(t, redis.Nil, err)
	assert.Equal(t, 0, i)
}
