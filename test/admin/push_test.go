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

