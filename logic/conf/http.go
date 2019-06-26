package conf

import (
	"github.com/spf13/viper"
	"gitlab.com/jetfueltw/cpw/micro/http"
)

func newHttp() *http.Conf {
	// TODO error 處理
	c, _ := http.ReadViper(viper.Sub("http"))
	return c
}

func newAdminHttp() *http.Conf {
	// TODO error 處理
	c, _ := http.ReadViper(viper.Sub("admin"))
	return c
}
