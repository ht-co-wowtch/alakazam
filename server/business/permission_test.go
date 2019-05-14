package business

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestIsBanned(t *testing.T) {
	assert.True(t, IsBanned(253))
}

func TestNotBanned(t *testing.T) {
	assert.False(t, IsBanned(255))
}