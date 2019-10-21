package models

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestRoomTableName(t *testing.T) {
	m := new(Room)

	assert.Equal(t, "rooms", m.TableName())
}

func TestCreateRoom(t *testing.T) {
	room := &Room{
		IsMessage:    true,
		IsBets:       true,
		DayLimit:     1,
		DepositLimit: 100,
		DmlLimit:     100,
	}

	aff, err := s.CreateRoom(room)

	assert.Nil(t, err)
	assert.Equal(t, int64(1), aff)

	r := new(Room)
	ok, err := x.Where("id = ?", room.Id).Get(r)

	assert.Nil(t, err)
	assert.True(t, ok)
	assert.True(t, r.IsMessage)
	assert.True(t, r.IsBets)
	assert.Equal(t, room.DayLimit, r.DayLimit)
	assert.Equal(t, room.DepositLimit, r.DepositLimit)
	assert.Equal(t, room.DmlLimit, r.DmlLimit)
}

func TestUpdateRoom(t *testing.T) {
	assert.NoError(t, prepareTestDatabase())

	room := Room{
		Id:           1,
		IsMessage:    false,
		IsBets:       false,
		DayLimit:     2,
		DepositLimit: 200,
		DmlLimit:     200,
	}

	aff, err := s.UpdateRoom(room)

	assert.Nil(t, err)
	assert.Equal(t, int64(1), aff)

	r := new(Room)
	ok, err := x.Where("id = ?", room.Id).Get(r)

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

	r, err := s.GetRoom(1)

	assert.Nil(t, err)
	assert.Equal(t, Room{
		IsMessage:    true,
		IsBets:       true,
		DayLimit:     1,
		DepositLimit: 100,
		DmlLimit:     1000,
		Status:       true,
		UpdateAt:     at,
		CreateAt:     at,
	}, r)
}

func TestGetChat(t *testing.T) {
	assert.NoError(t, prepareTestDatabase())

	room, message, err := s.GetChat(1)

	assert.Nil(t, err)
	assert.Equal(t, 1, room.Id)
	assert.Equal(t, int32(1), message.RoomId)
}

func TestGetChatNoMessage(t *testing.T) {
	assert.NoError(t, prepareTestDatabase())

	room, message, err := s.GetChat(2)

	assert.Nil(t, err)
	assert.Equal(t, 2, room.Id)
	assert.Equal(t, int32(0), message.RoomId)
}
