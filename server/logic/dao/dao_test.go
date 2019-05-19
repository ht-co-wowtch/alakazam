package dao

import (
	"gitlab.com/jetfueltw/cpw/alakazam/server/logic/conf"
	"os"
	"testing"
)

var (
	d *Dao
)

func TestMain(m *testing.M) {
	if err := conf.Read("../../../test/config/logic.yml"); err != nil {
		panic(err)
	}
	d = New(conf.Conf)
	if err := d.Ping(); err != nil {
		os.Exit(-1)
	}
	if err := d.Close(); err != nil {
		os.Exit(-1)
	}
	if err := d.Ping(); err == nil {
		os.Exit(-1)
	}
	d = New(conf.Conf)
	os.Exit(m.Run())
}
