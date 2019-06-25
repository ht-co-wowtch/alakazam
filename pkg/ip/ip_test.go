package ip

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestIpCheck(t *testing.T) {
	assert.Nil(t, Check("127.0.0.1"))
	assert.Nil(t, Check("127.0.0.1:8080"))
	assert.Error(t, Check("127.0.0.1:sdsdds"))
	assert.Error(t, Check("127.0.0.1:"))
	assert.Error(t, Check("127.0.0"))
}
