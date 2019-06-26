package permission

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBanned(t *testing.T) {
	assert.True(t, IsBanned(PlayDefaultPermission-Message))
}

func TestNotBanned(t *testing.T) {
	assert.False(t, IsBanned(PlayDefaultPermission))
}

func TestGetBonus(t *testing.T) {
	assert.True(t, IsGetBonus(getBonus+Message))
}

func TestNotGetBonus(t *testing.T) {
	assert.False(t, IsGetBonus(Message))
}

func TestSendFollow(t *testing.T) {
	assert.True(t, IsSendFollow(sendFollow+Message))
}

func TestNotSendFollow(t *testing.T) {
	assert.False(t, IsSendFollow(Message))
}

func TestGetFollow(t *testing.T) {
	assert.True(t, IsGetFollow(getFollow+Message))
}

func TestNotGetFollow(t *testing.T) {
	assert.False(t, IsGetFollow(Message))
}

func TestIsMoney(t *testing.T) {
	assert.True(t, IsMoney(money+look))
}

func TestNotIsMoney(t *testing.T) {
	assert.False(t, IsMoney(look))
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
