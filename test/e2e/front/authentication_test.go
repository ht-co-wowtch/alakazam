package front

import (
	"github.com/stretchr/testify/assert"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"gitlab.com/jetfueltw/cpw/alakazam/test/internal/request"
	"testing"
)

func TestNotAuthentication(t *testing.T) {
	r := request.PushRoomNotToken("", "", "")

	e := request.ToError(t, r.Body)
	e.Status = r.StatusCode

	assert.Equal(t, errors.AuthorizationError, e)
}
