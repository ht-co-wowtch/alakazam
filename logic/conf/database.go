package conf

import (
	"github.com/spf13/viper"
	"gitlab.com/jetfueltw/cpw/micro/database"
)

func newDatabase() *database.Conf {
	v := viper.Sub("database")

	// TODO error 處理
	c, _ := database.ReadViper(v)
	return c
}
