package cache

import (
	"github.com/stretchr/testify/assert"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/permission"
	"gitlab.com/jetfueltw/cpw/micro/id"
	"strconv"
	"testing"
	"time"
)

var (
	status = permission.PlayDefaultPermission
	day    = 1
	dml    = 100
	amount = 500
)

func TestSetRoom(t *testing.T) {
	roomId := id.UUid32()
	err := c.SetRoom(roomId, status, day, dml, amount)

	assert.Nil(t, err)

	m := r.HGetAll(keyRoom(roomId)).Val()

	assert.Equal(t, map[string]string{
		hashPermissionKey:  strconv.Itoa(status),
		hashLimitDayKey:    strconv.Itoa(day),
		hashLimitDmlKey:    strconv.Itoa(dml),
		hashLimitAmountKey: strconv.Itoa(amount),
	}, m)

	expire := r.TTL(keyRoom(roomId)).Val()

	assert.Equal(t, time.Hour, expire)
}

func TestGetRoomByMoney(t *testing.T) {
	roomId := id.UUid32()
	_ = c.SetRoom(roomId, status, day, dml, amount)

	dy, dl, a, err := c.GetRoomByMoney(roomId)

	assert.Nil(t, err)
	assert.Equal(t, day, dy)
	assert.Equal(t, dml, dl)
	assert.Equal(t, amount, a)
}

func TestGetRoom(t *testing.T) {
	roomId := id.UUid32()
	_ = c.SetRoom(roomId, status, day, dml, amount)

	s, err := c.GetRoom(roomId)

	assert.Nil(t, err)
	assert.Equal(t, s, s)
}
