package e2e

import (
	"github.com/stretchr/testify/assert"
	"gitlab.com/jetfueltw/cpw/alakazam/test/internal/request"
	"gitlab.com/jetfueltw/cpw/micro/errdefs"
	"gitlab.com/jetfueltw/cpw/micro/id"
	"net/http"
	"testing"
	"time"
)

func TestBanned(t *testing.T) {
	userA := request.DialAuth(t, id.UUid32(), uidA)
	userA.SetBanned("test", 5)
	r := userA.PushRoom("test")

	defer userA.DeleteBanned()

	e := r.Error.(*errdefs.Error)

	assert.Equal(t, http.StatusUnauthorized, r.StatusCode)
	assert.Equal(t, 10024013, e.Code)
	assert.Equal(t, "您在禁言状态，无法发言", e.Message)
}

func TestDeleteBanned(t *testing.T) {
	userA := request.DialAuth(t, id.UUid32(), uidA)
	userA.SetBanned("test", 60)
	userA.DeleteBanned()
	r := userA.PushRoom("test")

	assert.Equal(t, http.StatusNoContent, r.StatusCode)
}

func TestBannedExpired(t *testing.T) {
	userA := request.DialAuth(t, id.UUid32(), uidA)
	userA.SetBanned("test", 3)
	time.Sleep(time.Second * 4)
	r := userA.PushRoom("test")

	assert.Equal(t, http.StatusNoContent, r.StatusCode)
}
