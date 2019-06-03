package admin

import (
	"gitlab.com/jetfueltw/cpw/alakazam/logic/cache"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/conf"
	"gitlab.com/jetfueltw/cpw/alakazam/test/internal/run"
	"os"
	"testing"
)

var (
	d *cache.Cache
)

func TestMain(m *testing.M) {
	r := run.Run("../../config")

	if err := conf.Read("../../config/logic.yml"); err != nil {
		panic(err)
	}

	d = cache.NewRedis(conf.Conf.Redis)
	if err := d.Ping(); err != nil {
		os.Exit(-1)
	}
	if err := d.Close(); err != nil {
		os.Exit(-1)
	}
	if err := d.Ping(); err == nil {
		os.Exit(-1)
	}
	d = cache.NewRedis(conf.Conf.Redis)
	code := m.Run()
	r()
	os.Exit(code)
}
