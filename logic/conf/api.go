package conf

import "github.com/spf13/viper"

type Api struct {
	Host string
}

func newApi() *Api {
	return &Api{
		Host: viper.GetString("api.host"),
	}
}
