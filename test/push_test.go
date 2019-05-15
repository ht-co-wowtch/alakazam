package test

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"gitlab.com/jetfueltw/cpw/alakazam/server/errors"
	"gitlab.com/jetfueltw/cpw/alakazam/test/request"
	"net/http"
	"testing"
)

// 房間訊息推送成功
func TestPushRoom(t *testing.T) {
	a, err := request.DialAuth("1000")
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
	a, err := request.DialAuthToken("1000", "1")
	if err != nil {
		assert.Fail(t, err.Error())
		return
	}
	r := request.PushRoom(a.Uid, a.Key, "測試")

	e := new(errors.Error)
	json.Unmarshal(r.Body, e)

	assert.Equal(t, http.StatusUnauthorized, r.StatusCode)
	assert.Equal(t, 10024013, e.Code)
	assert.Equal(t, "您在禁言状态，无法发言", e.Message)
}
