package logic

import (
	"github.com/stretchr/testify/assert"
	"gitlab.com/jetfueltw/cpw/alakazam/server/logic/conf"
	"gitlab.com/jetfueltw/cpw/alakazam/server/logic/dao"
	"os"
	"testing"
)

var (
	l Logic
	d *dao.Dao
)

func TestMain(m *testing.M) {
	l, d = newTestDao()
	os.Exit(m.Run())
}

func newTestDao() (Logic, *dao.Dao) {
	initTestConfig()
	d := dao.New(conf.Conf)
	return Logic{c: conf.Conf, dao: d}, d
}

func initTestConfig() {
	if err := conf.Read("../../test/run/logic.yml"); err != nil {
		panic(err)
	}
}

func addUser(t *testing.T, uid, key, roomId, name string) {
	err := d.AddMapping(uid, key, roomId, name, "", 0)
	assert.Nil(t, err)
}
