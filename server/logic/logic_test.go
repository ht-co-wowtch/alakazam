package logic

import (
	"gitlab.com/jetfueltw/cpw/alakazam/server/logic/conf"
	"gitlab.com/jetfueltw/cpw/alakazam/server/logic/dao"
	test "gitlab.com/jetfueltw/cpw/alakazam/test/dao"
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

func addUser(t *testing.T, uid, key, roomId, name string, status int) {
	test.AddUser(d, t, uid, key, roomId, name, status)
}

func getUser(t *testing.T, uid, key string) (string, string, int) {
	return test.GetUser(d, t, uid, key)
}
