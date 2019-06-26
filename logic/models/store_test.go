package models

import (
	"fmt"
	"github.com/go-xorm/core"
	"github.com/go-xorm/xorm"
	_ "github.com/mattn/go-sqlite3"
	"gopkg.in/testfixtures.v2"
	"os"
	"testing"
	"time"
)

var (
	s        *Store
	x        *xorm.Engine
	fixtures *testfixtures.Context

	uidA = "1d7eff72ab49470882833853875340c1"
	uidB = "82ea16cd2d6a49d887440066ef739669"

	roomIdA = "6318a4f786e64c6487a30687e9df3a13"
)

func TestMain(m *testing.M) {
	var err error
	x, err = createTestEngine("./fixtures")
	db, err := xorm.NewEngineGroup(x, []*xorm.Engine{x})
	if err != nil {
		fatalTestError("Error creating test engine group: %v\n", err)
	}
	if err != nil {
		fatalTestError("Error creating test engine: %v\n", err)
	}
	s = &Store{
		d: db,
	}
	exitStatus := m.Run()
	db.Close()
	os.Exit(exitStatus)
}

// 建立測試用的xorm
func createTestEngine(fixturesDir string) (*xorm.Engine, error) {
	x, err := xorm.NewEngine("sqlite3", "file::memory:?cache=shared")
	if err != nil {
		return nil, err
	}
	x.SetMapper(core.GonicMapper{})
	x.SetTZDatabase(time.Local)
	if err = x.StoreEngine("InnoDB").Sync2(Table()...); err != nil {
		return nil, err
	}
	x.ShowSQL(true)
	return x, initFixtures(x, &testfixtures.SQLite{}, fixturesDir)
}

// 為測試數據庫初始化測試
func initFixtures(x *xorm.Engine, helper testfixtures.Helper, dir string) (err error) {
	testfixtures.SkipDatabaseNameCheck(true)
	fixtures, err = testfixtures.NewFolder(x.DB().DB, helper, dir)
	return err
}

func prepareTestDatabase() error {
	return fixtures.Load()
}

func fatalTestError(fmtStr string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, fmtStr, args...)
	os.Exit(1)
}
