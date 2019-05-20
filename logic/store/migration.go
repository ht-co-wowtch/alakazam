package store

import (
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/conf"
)

func RunMigration(path string) {
	runMigration(fmt.Sprintf("%s/sql", path))
}

func runMigration(path string) {
	db := NewDB(conf.Conf.DB)
	driver, _ := mysql.WithInstance(db, &mysql.Config{})
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
