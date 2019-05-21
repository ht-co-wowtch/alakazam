package request

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"gitlab.com/jetfueltw/cpw/alakazam/errors"
	"testing"
)

func ToError(t *testing.T, b [] byte) errors.Error {
	e := errors.Error{}
	err := json.Unmarshal(b, &e)
	assert.Nil(t, err)
	return e
}
