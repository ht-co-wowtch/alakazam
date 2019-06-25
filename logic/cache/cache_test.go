package cache

import (
	"fmt"
	"github.com/alicebob/miniredis"
	goRedis "github.com/go-redis/redis"
	"github.com/stretchr/testify/assert"
	"gitlab.com/jetfueltw/cpw/micro/redis"
	"os"
	"testing"
	"time"
)

var (
	r *goRedis.Client
	c *Cache
)

func TestMain(m *testing.M) {
	s, err := miniredis.Run()
	if err != nil {
		fatalTestError("Error creating redis test : %v\n", err)
	}
	r = redis.New(&redis.Conf{
		Addr: s.Addr(),
	})
	c = &Cache{
		c:      r,
		expire: time.Second * 10,
	}
	exitStatus := m.Run()
	s.Close()
	os.Exit(exitStatus)
}

func TestNewRedis(t *testing.T) {
	c := NewRedis(&redis.Conf{})

	assert.Equal(t, time.Minute*30, c.expire)
}

func fatalTestError(fmtStr string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, fmtStr, args...)
	os.Exit(1)
}
