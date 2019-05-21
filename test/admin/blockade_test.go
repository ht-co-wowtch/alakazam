package admin

import (
	"github.com/stretchr/testify/assert"
	"gitlab.com/jetfueltw/cpw/alakazam/test/internal/request"
	"net/http"
	"testing"
)

// 封鎖會員
func TestSetBlockade(t *testing.T) {
	request.DialAuth("1000")

	r := request.SetBlockade("82ea16cd2d6a49d887440066ef739669", "測試")
	assert.Nil(t, r.Error)
	assert.Equal(t, http.StatusNoContent, r.StatusCode)
	assert.Empty(t, string(r.Body))
}

// 解除某會員封鎖
func TestDeleteBlockade(t *testing.T) {
	request.DialAuth("1000")
	request.SetBlockade("82ea16cd2d6a49d887440066ef739669", "測試")

	r := request.DeleteBlockade("82ea16cd2d6a49d887440066ef739669")
	assert.Nil(t, r.Error)
	assert.Equal(t, http.StatusNoContent, r.StatusCode)
	assert.Empty(t, string(r.Body))
}
