package store

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
	"gitlab.com/jetfueltw/cpw/micro/database"
	"time"
)

var tables []interface{}

func init() {
	tables = append(tables,
		new(Member),
		new(Room),
	)
}

type Store struct {
	d *xorm.EngineGroup
}

func Table() []interface{} {
	return tables
}

func NewStore(c *database.Conf) *Store {
	// TODO 處理error
	d, _ := database.NewORM(c)
	d.SetTZDatabase(time.Local)
	return &Store{
		d: d,
	}
}
