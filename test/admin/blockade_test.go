package admin

import (
	"github.com/stretchr/testify/assert"
	"gitlab.com/jetfueltw/cpw/alakazam/test/internal/request"
	"net/http"
	"testing"
)

// 封鎖會員
func TestSetBlockade(t *testing.T) {
	a, err := request.DialAuth("4000")
	assert.Nil(t, err)

	r := request.SetBlockade(a.Uid, "測試")
	assert.Nil(t, r.Error)
	assert.Equal(t, http.StatusNoContent, r.StatusCode)
	assert.Empty(t, string(r.Body))
}

// 解除某會員封鎖
func TestDeleteBlockade(t *testing.T) {
	a, err := request.DialAuth("4001")
	assert.Nil(t, err)

	request.SetBlockade(a.Uid, "測試")

	r := request.DeleteBlockade(a.Uid)
	assert.Nil(t, r.Error)
	assert.Equal(t, http.StatusNoContent, r.StatusCode)
	assert.Empty(t, string(r.Body))
}
