package logic

import (
	"context"
	"gitlab.com/jetfueltw/cpw/alakazam/internal/logic/conf"
	"os"
	"testing"
)

var (
	lg *Logic
)

func TestMain(m *testing.M) {
	if err := conf.Read("../../cmd/logic/logic-example.yml"); err != nil {
		panic(err)
	}
	lg = New(conf.Conf)
	if err := lg.Ping(context.TODO()); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}
