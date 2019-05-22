package conf

import (
	"github.com/spf13/viper"
	"time"
)

// http server config
type HTTPServer struct {
	// port
	Addr string

	// 沒用到
	ReadTimeout time.Duration

	// 沒用到
	WriteTimeout time.Duration

	Cors Cors
}

type Cors struct {
	Origins []string

	Headers []string
}

func newHttp() *HTTPServer {
	return &HTTPServer{
		Addr:         viper.GetString("httpServer.host"),
		ReadTimeout:  time.Duration(viper.GetInt("httpServer.readTimeout")) * time.Second,
		WriteTimeout: time.Duration(viper.GetInt("httpServer.writeTimeout")) * time.Second,
		Cors: Cors{
			Origins: viper.GetStringSlice("httpServer.cors.origins"),
			Headers: viper.GetStringSlice("httpServer.cors.headers"),
		},
	}
}

func newAdminHttp() *HTTPServer {
	return &HTTPServer{
		Addr:         viper.GetString("httpAdminServer.host"),
		ReadTimeout:  time.Duration(viper.GetInt("httpAdminServer.readTimeout")) * time.Second,
		WriteTimeout: time.Duration(viper.GetInt("httpAdminServer.writeTimeout")) * time.Second,
	}
}
