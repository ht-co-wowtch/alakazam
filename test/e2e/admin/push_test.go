package admin

import (
	"github.com/stretchr/testify/assert"
	"gitlab.com/jetfueltw/cpw/alakazam/test/internal/request"
	"net/http"
	"testing"
)

// 廣播訊息推送
func TestPushBroadcast(t *testing.T) {
	_, err := request.DialAuth("6000")
	if err != nil {
		assert.Fail(t, err.Error())
		return
	}
	r := request.PushBroadcast([]string{"6000", "6001"}, "測試")

	assert.Equal(t, http.StatusNoContent, r.StatusCode)
	assert.Empty(t, r.Body)
}
