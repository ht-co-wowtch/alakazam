package e2e

import (
	"gitlab.com/jetfueltw/cpw/alakazam/test/internal/run"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	r := run.Run("../config")
	i := m.Run()
	r()
	os.Exit(i)
}
