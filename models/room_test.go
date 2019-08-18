package models

import (
	"github.com/stretchr/testify/assert"
	"gitlab.com/jetfueltw/cpw/micro/id"
	"testing"
	"time"
)

func TestRoomTableName(t *testing.T) {
	m := new(Room)

	assert.Equal(t, "rooms", m.TableName())
}

func TestCreateRoom(t *testing.T) {
	room := Room{
		Uuid:         id.UUid32(),
		IsMessage:    true,
		IsFollow:     true,
		DayLimit:     1,
		DepositLimit: 100,
		DmlLimit:     100,
	}

	aff, err := s.CreateRoom(room)

	assert.Nil(t, err)
	assert.Equal(t, int64(1), aff)

	r := new(Room)
	ok, err := x.Where("uuid = ?", room.Uuid).Get(r)

	assert.Nil(t, err)
	assert.True(t, ok)
	assert.True(t, r.IsMessage)
	assert.True(t, r.IsFollow)
	assert.Equal(t, room.DayLimit, r.DayLimit)
	assert.Equal(t, room.DepositLimit, r.DepositLimit)
	assert.Equal(t, room.DmlLimit, r.DmlLimit)
}

func TestUpdateRoom(t *testing.T) {
	assert.NoError(t, prepareTestDatabase())

	room := Room{
		Uuid:         roomIdA,
		IsMessage:    false,
		IsFollow:     false,
		DayLimit:     2,
		DepositLimit: 200,
		DmlLimit:     200,
	}

	aff, err := s.UpdateRoom(room)

	assert.Nil(t, err)
	assert.Equal(t, int64(1), aff)

	r := new(Room)
	ok, err := x.Where("id = ?", roomIdA).Get(r)

	assert.Nil(t, err)
	assert.True(t, ok)
	assert.False(t, r.IsMessage)
	assert.Equal(t, room.DayLimit, r.DayLimit)
	assert.Equal(t, room.DepositLimit, r.DepositLimit)
	assert.Equal(t, room.DmlLimit, r.DmlLimit)
}

func TestGetRoom(t *testing.T) {
	assert.NoError(t, prepareTestDatabase())

	at, _ := time.ParseInLocation("2006-01-02 15:04:05", "2019-06-26 13:52:32", time.Local)

	r, ok, err := s.GetRoom(roomIdA)

	assert.Nil(t, err)
	assert.True(t, ok)
	assert.Equal(t, Room{
		Uuid:         roomIdA,
		IsMessage:    true,
		IsFollow:     true,
		DayLimit:     1,
		DepositLimit: 100,
		DmlLimit:     1000,
		Status:       true,
		UpdateAt:     at,
		CreateAt:     at,
	}, r)
}
