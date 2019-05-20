package logic

import (
	"github.com/stretchr/testify/assert"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/business"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/cache"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/conf"
	"testing"
)

func TestIsNotBanned(t *testing.T) {
	l := Logic{}
	is := l.isBanned("123", 3)
	assert.False(t, is)
}

func TestIsBanned(t *testing.T) {
	err := d.SetBanned("456", 10)
	assert.Nil(t, err)

	is := l.isBanned("456", 1)
	assert.True(t, is)
}

func TestBannedExpire(t *testing.T) {
	addUser(t, "789", "1", "", "", 1)

	is := l.isBanned("789", 1)
	assert.False(t, is)

	_, _, s := getUser(t, "789", "1")
	assert.Equal(t, 1+business.Message, s)
}

func TestBannedError(t *testing.T) {
	initTestConfig()
	conf.Conf.Redis.Addr = ":1111"
	l := Logic{cache: cache.NewRedis(conf.Conf.Redis)}

	is := l.isBanned("", 1)
	assert.False(t, is)
}
