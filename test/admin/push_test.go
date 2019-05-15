package admin

import (
	"github.com/stretchr/testify/assert"
	"gitlab.com/jetfueltw/cpw/alakazam/test/request"
	"gitlab.com/jetfueltw/cpw/alakazam/test/run"
	"net/http"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	r := run.Run("../run")
	defer r()
	os.Exit(m.Run())
}

// 廣播訊息推送
func TestPushBroadcast(t *testing.T) {
	a, err := request.DialAuth("1000")
	if err != nil {
		assert.Fail(t, err.Error())
		return
	}
	r := request.PushBroadcast(a.Uid, a.Key, "測試", []string{"1000", "1001"})

	assert.Equal(t, http.StatusNoContent, r.StatusCode)
	assert.Empty(t, r.Body)
}

// 設定某會員禁言5秒
func TestSetBanned(t *testing.T) {
	r := request.SetBanned("82ea16cd2d6a49d887440066ef739669", "測試", 5)
	assert.Nil(t, r.Error)
	assert.Equal(t, http.StatusNoContent, r.StatusCode)
	assert.Empty(t, string(r.Body))
}
