package store

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/conf"
	"time"
)

type Store struct {
	*sql.DB
}

func NewStore(c *conf.Database) *Store {
	return &Store{NewDB(c)}
}

func NewDB(c *conf.Database) *sql.DB {
	db, err := sql.Open(c.Driver, DatabaseDns(c))
	if err != nil {
		panic(err)
	}

	db.SetMaxOpenConns(c.MaxOpenConn)
	db.SetMaxIdleConns(c.MaxIdleConn)
	db.SetConnMaxLifetime(time.Duration(c.ConnMaxLifetime) * time.Second)

	if err := db.Ping(); err != nil {
		panic(err)
	}
	return db
}

func DatabaseDns(c *conf.Database) string {
	return fmt.Sprintf("%v:%v@tcp(%v:%v)/%v?charset=%v&collation=%v&parseTime=true&timeout=2s&loc=Local", c.User, c.Password, c.Host, c.Port, c.Database, c.Charset, c.Collation)
}
