package admin

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/store"
	"gitlab.com/jetfueltw/cpw/alakazam/test/internal/request"
	"net/http"
	"testing"
)

var room struct {
	RoomId string `json:"room_id"`
}

func TestGetRoomByEmpty(t *testing.T) {
	r := request.GetRoom("580d209be6b043f2a992518db5e7269d")
	e := request.ToError(t, r.Body)

	assert.Nil(t, r.Error)
	assert.Equal(t, errors.NoRowsError.Code, e.Code)
	assert.Equal(t, errors.NoRowsError.Message, e.Message)
}

func TestCreateRoom(t *testing.T) {
	r := request.CreateRoom(store.Room{
		IsMessage: true,
	})

	json.Unmarshal(r.Body, &room)

	assert.Equal(t, http.StatusOK, r.StatusCode)
	assert.NotEmpty(t, room.RoomId)
	assert.Len(t, room.RoomId, 32)

	r = request.GetRoom(room.RoomId)

	room := new(store.Room)
	json.Unmarshal(r.Body, room)

	assert.Equal(t, &store.Room{
		IsMessage: true,
	}, room)
}

func TestUpdateRoom(t *testing.T) {
	r := request.CreateRoom(store.Room{
		IsMessage: true,
	})

	json.Unmarshal(r.Body, &room)

	r = request.UpdateRoom(room.RoomId, store.Room{})

	assert.Equal(t, http.StatusNoContent, r.StatusCode)

	r = request.GetRoom(room.RoomId)

	room := new(store.Room)
	json.Unmarshal(r.Body, room)

	assert.Equal(t, &store.Room{}, room)
}

func TestSetRoomDayByDayNotEmpty(t *testing.T) {
	r := request.CreateRoom(store.Room{
		Limit: store.Limit{
			Day: 1,
		},
	})

	e := request.ToError(t, r.Body)

	assert.Equal(t, http.StatusUnprocessableEntity, r.StatusCode)
	assert.Equal(t, errors.SetRoomError.Code, e.Code)
	assert.Equal(t, "储值或打码量不可都小于等于0", e.Message)
}

func TestSetRoomDayByDayEmpty(t *testing.T) {
	r := request.CreateRoom(store.Room{
		Limit: store.Limit{
			Dml: 1,
		},
	})

	e := request.ToError(t, r.Body)

	assert.Equal(t, errors.SetRoomError.Code, e.Code)
	assert.Equal(t, "储值跟打码量都需是0", e.Message)
}

func TestSetRoomDayByDayLimit(t *testing.T) {
	r := request.CreateRoom(store.Room{
		Limit: store.Limit{
			Day:    32,
			Amount: 1000,
		},
	})

	e := request.ToError(t, r.Body)

	assert.Equal(t, errors.SetRoomError.Code, e.Code)
	assert.Equal(t, "储值跟打码量聊天限制天数不能大于31", e.Message)
}

func TestSetRoomDayByDayNegative(t *testing.T) {
	r := request.CreateRoom(store.Room{
		Limit: store.Limit{
			Day: -1,
		},
	})

	e := request.ToError(t, r.Body)

	assert.Equal(t, errors.FailureError.Code, e.Code)
	assert.Equal(t, errors.FailureError.Message, e.Message)
}
