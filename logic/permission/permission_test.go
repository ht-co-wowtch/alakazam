package permission

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsBanned(t *testing.T) {
	assert.True(t, IsBanned(253))
}

func TestNotBanned(t *testing.T) {
	assert.False(t, IsBanned(255))
}

func TestIsLook(t *testing.T) {
	assert.True(t, IsLook(255))
}

func TestNotLook(t *testing.T) {
	assert.False(t, IsLook(6))
}

func TestIsSendBonus(t *testing.T) {
	assert.True(t, IsSendBonus(4))
}

func TestNotSendBonus(t *testing.T) {
	assert.False(t, IsSendBonus(3))
}

func TestIsGetBonus(t *testing.T) {
	assert.True(t, IsGetBonus(9))
}

func TestNotGetBonus(t *testing.T) {
	assert.False(t, IsGetBonus(3))
}

func TestIsSendFollow(t *testing.T) {
	assert.True(t, IsSendFollow(17))
}

func TestNotSendFollow(t *testing.T) {
	assert.False(t, IsSendFollow(3))
}

func TestIsGetFollow(t *testing.T) {
	assert.True(t, IsGetFollow(33))
}

func TestNotGetFollow(t *testing.T) {
	assert.False(t, IsGetFollow(3))
}

func TestNewPermission(t *testing.T) {
	p := NewPermission(253)
	assert.Equal(t, &Permission{
		Message:    false,
		SendFollow: true,
		GetFollow:  true,
		SendBonus:  true,
		GetBonus:   true,
	}, p)
}
