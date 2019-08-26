package message

import (
	"fmt"
	"github.com/alicebob/miniredis"
	goRedis "github.com/go-redis/redis"
	"gitlab.com/jetfueltw/cpw/micro/redis"
	"os"
	"testing"
)

var (
	r *goRedis.Client
)

func TestMain(m *testing.M) {
	s, err := miniredis.Run()
	if err != nil {
		fatalTestError("Error creating redis test : %v\n", err)
	}
	r = redis.New(&redis.Conf{
		Addr: s.Addr(),
	})
	exitStatus := m.Run()
	s.Close()
	os.Exit(exitStatus)
}

func fatalTestError(fmtStr string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, fmtStr, args...)
	os.Exit(1)
}
