package test

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"gitlab.com/jetfueltw/cpw/alakazam/test/internal/request"
	"net/http"
	"testing"
	"time"
)

// 房間訊息推送成功
func TestPushRoom(t *testing.T) {
	a, err := request.DialAuth("2000")
	if err != nil {
		assert.Fail(t, err.Error())
		return
	}
	r := request.PushRoom(a.Uid, a.Key, "測試")

	assert.Equal(t, http.StatusNoContent, r.StatusCode)
	assert.Empty(t, r.Body)
}

// 禁言
func TestPushRoomBanned(t *testing.T) {
	a := givenBanned(t, "2001")
	time.Sleep(time.Second * 4)
	r := request.PushRoom(a.Uid, a.Key, "測試")

	assert.Equal(t, http.StatusNoContent, r.StatusCode)
}

// 解除禁言
func TestPushRoomRemoveBanned(t *testing.T) {
	a := givenBanned(t, "2002")
	request.DeleteBanned(a.Uid)
	r := request.PushRoom(a.Uid, a.Key, "測試")

	assert.Equal(t, http.StatusNoContent, r.StatusCode)
}

func givenBanned(t *testing.T, roomId string) request.Auth {
	a, err := request.DialAuth(roomId)
	if err != nil {
		assert.Fail(t, err.Error())
		return a
	}

	request.SetBanned(a.Uid, "測試", 3)
	r := request.PushRoom(a.Uid, a.Key, "測試")

	e := new(errors.Error)
	json.Unmarshal(r.Body, e)

	assert.Equal(t, http.StatusUnauthorized, r.StatusCode)
	assert.Equal(t, 10024013, e.Code)
	assert.Equal(t, "您在禁言状态，无法发言", e.Message)
	return a
}
