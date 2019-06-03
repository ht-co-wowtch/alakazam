package admin

import (
	"github.com/stretchr/testify/assert"
	"gitlab.com/jetfueltw/cpw/alakazam/test/internal/request"
	"net/http"
	"testing"
	"time"
)

// 設定某會員禁言5秒
func TestSetBanned(t *testing.T) {
	uid := "00ea16cd2d6a49d887440066ef739669"
	expectedUnix := time.Now().Add(time.Second * time.Duration(5)).Unix()
	r := request.SetBanned(uid, "測試", 5)

	assert.Nil(t, r.Error)
	assert.Equal(t, http.StatusNoContent, r.StatusCode)
	assert.Empty(t, string(r.Body))

	ex, isBanned, _ := d.GetBanned(uid)
	assert.True(t, isBanned)
	assert.Equal(t, expectedUnix, ex.Unix())
}

// 解除某會員禁言
func TestDeleteBanned(t *testing.T) {
	uid := "11ea16cd2d6a49d887440066ef739669"
	request.SetBanned(uid, "測試", 5)
	r := request.DeleteBanned(uid)

	assert.Nil(t, r.Error)
	assert.Equal(t, http.StatusNoContent, r.StatusCode)
	assert.Empty(t, string(r.Body))

	_, isBanned, _ := d.GetBanned(uid)
	assert.False(t, isBanned)
}
