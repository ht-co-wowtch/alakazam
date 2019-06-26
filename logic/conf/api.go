package conf

import (
	"github.com/spf13/viper"
	"gitlab.com/jetfueltw/cpw/micro/client"
)

func newApi() *client.Conf {
	v := viper.Sub("api")
	// TODO 處理error
	c, _ := client.ReadViper(v)
	return c
}
