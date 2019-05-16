package admin

import (
	"github.com/stretchr/testify/assert"
	"gitlab.com/jetfueltw/cpw/alakazam/test/request"
	"net/http"
	"testing"
)

// 設定某會員禁言5秒
func TestSetBanned(t *testing.T) {
	r := request.SetBanned("82ea16cd2d6a49d887440066ef739669", "測試", 5)
	assert.Nil(t, r.Error)
	assert.Equal(t, http.StatusNoContent, r.StatusCode)
	assert.Empty(t, string(r.Body))
}

