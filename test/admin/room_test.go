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

func TestGetRoomByEmpty(t *testing.T) {
	r := request.GetRoom(1000)
	e := request.ToError(t, r.Body)

	assert.Nil(t, r.Error)
	assert.Equal(t, errors.NoRowsError.Code, e.Code)
	assert.Equal(t, errors.NoRowsError.Message, e.Message)
}

func TestSetRoom(t *testing.T) {
	r := request.SetRoom(store.Room{
		RoomId:    1000,
		IsMessage: true,
	})
	assert.Nil(t, r.Error)
	assert.Empty(t, string(r.Body))

	r = request.GetRoom(1000)

	room := new(store.Room)
	json.Unmarshal(r.Body, room)

	assert.Equal(t, &store.Room{
		RoomId:    1000,
		IsMessage: true,
	}, room)
}

func TestSetRoomNotId(t *testing.T) {
	r := request.SetRoom(store.Room{})

	e := request.ToError(t, r.Body)

	assert.Equal(t, http.StatusUnprocessableEntity, r.StatusCode)
	assert.Equal(t, errors.DataError.Code, e.Code)
	assert.Equal(t, errors.DataError.Message, e.Message)
}

func TestSetRoomDayByDayNotEmpty(t *testing.T) {
	r := request.SetRoom(store.Room{
		RoomId: 1000,
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
	r := request.SetRoom(store.Room{
		RoomId: 1000,
		Limit: store.Limit{
			Dml: 1,
		},
	})

	e := request.ToError(t, r.Body)

	assert.Equal(t, errors.SetRoomError.Code, e.Code)
	assert.Equal(t, "储值跟打码量都需是0", e.Message)
}

func TestSetRoomDayByDayLimit(t *testing.T) {
	r := request.SetRoom(store.Room{
		RoomId: 1000,
		Limit: store.Limit{
			Day:    31,
			Amount: 1000,
		},
	})

	e := request.ToError(t, r.Body)

	assert.Equal(t, errors.SetRoomError.Code, e.Code)
	assert.Equal(t, "储值跟打码量聊天限制天数不能大于30", e.Message)
}

func TestSetRoomDayByDayNegative(t *testing.T) {
	r := request.SetRoom(store.Room{
		RoomId: 1000,
		Limit: store.Limit{
			Day: -1,
		},
	})

	e := request.ToError(t, r.Body)

	assert.Equal(t, errors.FailureError.Code, e.Code)
	assert.Equal(t, errors.FailureError.Message, e.Message)
}
