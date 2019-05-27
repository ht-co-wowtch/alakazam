package front

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
	Id string `json:"room_id"`
}

func TestRoomIsBanned(t *testing.T) {
	r := request.CreateRoom(store.Room{
		IsMessage: false,
	})

	json.Unmarshal(r.Body, &room)

	a, _ := request.DialAuth(room.Id)

	r = request.PushRoom(a.Uid, a.Key, "測試")
	e := request.ToError(t, r.Body)

	assert.Equal(t, errors.RoomBannedError.Code, e.Code)
	assert.Equal(t, errors.RoomBannedError.Message, e.Message)
}

func TestRoomRemoveBanned(t *testing.T) {
	r := request.CreateRoom(store.Room{
		IsMessage: false,
	})

	json.Unmarshal(r.Body, &room)

	a, _ := request.DialAuth(room.Id)
	request.PushRoom(a.Uid, a.Key, "測試")

	request.UpdateRoom(room.Id, store.Room{
		IsMessage: true,
	})
	r = request.PushRoom(a.Uid, a.Key, "測試")

	assert.Equal(t, http.StatusNoContent, r.StatusCode)
}

func TestRoomSetBanned(t *testing.T) {
	r := request.CreateRoom(store.Room{
		IsMessage: true,
	})

	json.Unmarshal(r.Body, &room)

	a, _ := request.DialAuth(room.Id)
	request.PushRoom(a.Uid, a.Key, "測試")

	request.UpdateRoom(room.Id, store.Room{
		IsMessage: false,
	})
	r = request.PushRoom(a.Uid, a.Key, "測試")
	e := request.ToError(t, r.Body)

	assert.Equal(t, errors.RoomBannedError.Code, e.Code)
}
