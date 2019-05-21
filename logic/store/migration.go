package store

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/conf"
)

func RunMigration(path string) {
	runMigration(fmt.Sprintf("%s/sql", path))
}

func runMigration(path string) {
	db := NewDB(conf.Conf.DB)
	driver, err := mysql.WithInstance(db, &mysql.Config{})
	if err != nil {
		panic(err)
	}
	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", path),
		"mysql",
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
