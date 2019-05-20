package dao

import (
	"gitlab.com/jetfueltw/cpw/alakazam/server/logic/conf"
	"os"
	"testing"
)

var (
	d *Cache
)

func TestMain(m *testing.M) {
	if err := conf.Read("../../../test/config/logic.yml"); err != nil {
		panic(err)
	}
	d = NewRedis(conf.Conf.Redis)
	if err := d.Ping(); err != nil {
		os.Exit(-1)
	}
	if err := d.Close(); err != nil {
		os.Exit(-1)
	}
	if err := d.Ping(); err == nil {
		os.Exit(-1)
	}
	d = NewRedis(conf.Conf.Redis)
	os.Exit(m.Run())
}
