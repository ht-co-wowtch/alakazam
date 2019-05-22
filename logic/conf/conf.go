package conf

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/spf13/viper"
)

var (
	Conf *Config
)

type Config struct {
	RPCServer       *RPCServer
	HTTPServer      *HTTPServer
	HTTPAdminServer *HTTPServer
	DB              *Database
	Kafka           *Kafka
	Redis           *Redis

	// comet連線用戶心跳，server會清除在線紀錄
	Heartbeat int64
}

func Read(path string) (err error) {
	viper.SetConfigType("yaml")
	var b []byte
	b, err = ioutil.ReadFile(path)
	if err != nil {
		return
	}
	if err = viper.ReadConfig(bytes.NewBuffer(b)); err != nil {
		return
	} else {
		fmt.Println("Using config file:", path)
	}
	Conf = load()
	if Conf.Heartbeat >= Conf.Redis.Expire.Nanoseconds() {
		return fmt.Errorf("comet心跳不能比redis expire還大")
	}
	return
}

// 載入config
func load() *Config {
	return &Config{
		DB:              newDatabase(),
		RPCServer:       newRpc(),
		HTTPServer:      newHttp(),
		HTTPAdminServer: newAdminHttp(),
		Redis:           newRedis(),
		Kafka:           newKafka(),
		Heartbeat:       (time.Duration(viper.GetInt("heartbeat")) * time.Second).Nanoseconds(),
	}
}
