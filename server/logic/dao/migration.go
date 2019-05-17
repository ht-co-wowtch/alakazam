package dao

import (
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"gitlab.com/jetfueltw/cpw/alakazam/server/logic/conf"
)

func RunMigration(path string) {
	runMigration(fmt.Sprintf("%s/sql", path))
}

func runMigration(path string) {
	db := newDB(conf.Conf.DB)
	driver, _ := mysql.WithInstance(db, &mysql.Config{})
	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", path),
		conf.Conf.DB.Driver,
		driver,
	)

	if err != nil {
		panic(err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		panic(fmt.Sprintf("An error occurred while syncing the database.. %v", err))
	}
	fmt.Println("Database migrated")
}
