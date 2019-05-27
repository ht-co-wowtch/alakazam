package store

import (
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/conf"
	"os"
	"testing"
)

var (
	store *Store
	mock  sqlmock.Sqlmock
)

func TestMain(m *testing.M) {
	d, mo, err := sqlmock.New()
	if err != nil {
		panic(err)
	}
	store = &Store{d}
	mock = mo
	os.Exit(m.Run())
}

func TestDatabaseDns(t *testing.T) {
	c := &conf.Database{
		Host:      "127.0.0.1",
		Port:      ":3306",
		Database:  "database",
		User:      "root",
		Password:  "root",
		Charset:   "utf8mb4",
		Collation: "utf8mb4_unicode_ci",
	}
	dns := DatabaseDns(c)

	assert.Equal(t, "root:root@tcp(127.0.0.1::3306)/database?charset=utf8mb4&collation=utf8mb4_unicode_ci&parseTime=true&timeout=2s&loc=Local", dns)
}
