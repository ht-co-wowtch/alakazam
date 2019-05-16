package admin

import (
	"gitlab.com/jetfueltw/cpw/alakazam/test/run"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	r := run.Run("../run")
	defer r()
	os.Exit(m.Run())
}
