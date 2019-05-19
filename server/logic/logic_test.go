package logic

import (
	"github.com/DATA-DOG/go-sqlmock"
	"gitlab.com/jetfueltw/cpw/alakazam/server/logic/conf"
	"gitlab.com/jetfueltw/cpw/alakazam/server/logic/dao"
	test "gitlab.com/jetfueltw/cpw/alakazam/test/dao"
	"os"
	"testing"
)

var (
	l      Logic
	d      *dao.Cache
	mockDB sqlmock.Sqlmock
)

func TestMain(m *testing.M) {
	l, d = newTestDao()
	os.Exit(m.Run())
}

func newTestDao() (Logic, *dao.Cache) {
	initTestConfig()
	c := dao.NewRedis(conf.Conf.Redis)
	db, mock, err := sqlmock.New()
	if err != nil {
		panic(err)
	}
	mockDB = mock
	return Logic{
		c:     conf.Conf,
		cache: dao.NewRedis(conf.Conf.Redis),
		db:    &dao.Store{db},
	}, c
}

func initTestConfig() {
	if err := conf.Read("../../test/config/logic.yml"); err != nil {
		panic(err)
	}
}

func addUser(t *testing.T, uid, key, roomId, name string, status int) {
	test.AddUser(d, t, uid, key, roomId, name, status)
}

func getUser(t *testing.T, uid, key string) (string, string, int) {
	return test.GetUser(d, t, uid, key)
}
