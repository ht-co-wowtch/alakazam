package e2e

import (
	"github.com/gin-gonic/gin"
	"gitlab.com/jetfueltw/cpw/alakazam/test/internal/run"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.ReleaseMode)
	r := run.Run("../config")
	i := m.Run()
	r()
	os.Exit(i)
}
