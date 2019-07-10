package conf

import (
	"gitlab.com/jetfueltw/cpw/micro/client"
	"gitlab.com/jetfueltw/cpw/micro/config"
	"gitlab.com/jetfueltw/cpw/micro/database"
	"gitlab.com/jetfueltw/cpw/micro/grpc"
	"gitlab.com/jetfueltw/cpw/micro/http"
	"gitlab.com/jetfueltw/cpw/micro/log"
	"gitlab.com/jetfueltw/cpw/micro/redis"
	"time"

	"github.com/spf13/viper"
)

var (
	Conf *Config
)

type Config struct {
	RPCServer  *grpc.Conf
	HTTPServer *http.Conf
	DB         *database.Conf
	Kafka      *Kafka
	Redis      *redis.Conf
	Api        *client.Conf
	// comet連線用戶心跳，server會清除在線紀錄
	Heartbeat int64
}

// Kafka
type Kafka struct {
	// Kafka 推送與接收Topic
	Topic string

	// 節點ip
	Brokers []string
}

func init() {
	config.SetEnvReplace(true)
	config.SetEnvPrefix("alakazam")
}

func Read(path string) error {
	v, err := config.Read(path)
	if err != nil {
		return err
	}

	Conf = new(Config)
	Conf.HTTPServer, err = http.ReadViper(v.Sub("http"))
	if err != nil {
		return err
	}
	Conf.RPCServer, err = grpc.ReadViper(v.Sub("grpcServer"))
	if err != nil {
		return err
	}
	Conf.Api, err = client.ReadViper(v.Sub("api"))
	if err != nil {
		return err
	}
	Conf.Redis, err = redis.ReadViper(v.Sub("redis"))
	if err != nil {
		return err
	}
	Conf.DB, err = database.ReadViper(v.Sub("database"))
	if err != nil {
		return err
	}
	Conf.Kafka = &Kafka{
		Topic:   v.GetString("kafka.topic"),
		Brokers: v.GetStringSlice("kafka.brokers"),
	}
	Conf.Heartbeat = (time.Duration(viper.GetInt("heartbeat")) * time.Second).Nanoseconds()
	l, err := log.ReadViper(v.Sub("log"))
	if err != nil {
		return err
	}
	return log.Start(l)
}
