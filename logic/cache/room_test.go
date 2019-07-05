package cache

import (
	"github.com/stretchr/testify/assert"
	"gitlab.com/jetfueltw/cpw/alakazam/models"
	"gitlab.com/jetfueltw/cpw/micro/id"
	"strconv"
	"testing"
	"time"
)

var (
	day    = 1
	dml    = 100
	amount = 500

	room = models.Room{
		Id:           id.UUid32(),
		IsMessage:    true,
		DayLimit:     day,
		DmlLimit:     dml,
		DepositLimit: amount,
	}
)

func TestSetRoom(t *testing.T) {
	err := c.SetRoom(room)

	assert.Nil(t, err)

	m := r.HGetAll(keyRoom(room.Id)).Val()

	assert.Equal(t, map[string]string{
		hashPermissionKey:  "5",
		hashLimitDayKey:    strconv.Itoa(day),
		hashLimitDmlKey:    strconv.Itoa(dml),
		hashLimitAmountKey: strconv.Itoa(amount),
	}, m)

	expire := r.TTL(keyRoom(room.Id)).Val()

	assert.Equal(t, time.Hour, expire)
}

func TestGetRoomByMoney(t *testing.T) {
	_ = c.SetRoom(room)

	dy, dl, a, err := c.GetRoomByMoney(room.Id)

	assert.Nil(t, err)
	assert.Equal(t, day, dy)
	assert.Equal(t, dml, dl)
	assert.Equal(t, amount, a)
}

func TestGetRoom(t *testing.T) {
	_ = c.SetRoom(room)

	s, err := c.GetRoom(room.Id)

	assert.Nil(t, err)
	assert.Equal(t, s, s)
}
