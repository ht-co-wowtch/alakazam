package admin

import (
	"gitlab.com/jetfueltw/cpw/alakazam/test/internal/run"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	r := run.Run("../config")
	defer r()
	os.Exit(m.Run())
}
