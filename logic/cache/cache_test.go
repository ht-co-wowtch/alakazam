package cache

import (
	"github.com/gomodule/redigo/redis"
	"github.com/rafaeljusto/redigomock"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/conf"
	"os"
	"testing"
)

var (
	d    *Cache
	mock *redigomock.Conn
)

func TestMain(m *testing.M) {
	mock = redigomock.NewConn()
	d = NewRedisDial(new(conf.Redis), func() (conn redis.Conn, e error) {
		return mock, nil
	})
	os.Exit(m.Run())
}