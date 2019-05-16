package logic

import (
	"github.com/stretchr/testify/assert"
	"gitlab.com/jetfueltw/cpw/alakazam/server/logic/business"
	"gitlab.com/jetfueltw/cpw/alakazam/server/logic/conf"
	"gitlab.com/jetfueltw/cpw/alakazam/server/logic/dao"
	"testing"
)

func TestIsNotBanned(t *testing.T) {
	l := Logic{}
	is := l.isBanned("123", 3)
	assert.False(t, is)
}

func TestIsBanned(t *testing.T) {
	l, d := newTestDao()

	err := d.SetBanned("456", 10)
	assert.Nil(t, err)

	is := l.isBanned("456", 1)
	assert.True(t, is)
}

func TestBannedExpire(t *testing.T) {
	l, d := newTestDao()

	err := d.AddMapping("789", "1", "", "", "", 1)
	assert.Nil(t, err)

	is := l.isBanned("789", 1)
	assert.False(t, is)

	_, _, s, _ := d.UserData("789", "1")
	assert.Equal(t, 1+business.Message, s)
}

func TestBannedError(t *testing.T) {
	initTestConfig()
	conf.Conf.Redis.Addr = ":1111"
	d := dao.New(conf.Conf)
	l := Logic{dao: d}

	is := l.isBanned("", 1)
	assert.False(t, is)
}

func newTestDao() (Logic, *dao.Dao) {
	initTestConfig()
	d := dao.New(conf.Conf)
	return Logic{dao: d}, d
}

func initTestConfig() {
	if err := conf.Read("../../test/run/logic.yml"); err != nil {
		panic(err)
	}
}
