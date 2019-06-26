package store

import (
	"fmt"
	"gitlab.com/jetfueltw/cpw/micro/database"
)

func Migrate(conf *database.Conf) error {
	x, err := database.NewORM(conf)
	if err != nil {
		return err
	}
	if err := x.Sync2(Table()...); err != nil {
		return err
	}
	fmt.Println("migrate success")
	return nil
}
