package admin

import (
	"github.com/stretchr/testify/assert"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/store"
	"gitlab.com/jetfueltw/cpw/alakazam/test/internal/request"
	"net/http"
	"testing"
)

func TestSetRoom(t *testing.T) {
	r := request.SetRoom(store.Room{
		RoomId: 1000,
	})
	assert.Nil(t, r.Error)
	assert.Equal(t, http.StatusNoContent, r.StatusCode)
	assert.Empty(t, string(r.Body))
}
