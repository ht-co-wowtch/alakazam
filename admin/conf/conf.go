package conf

import (
	"gitlab.com/jetfueltw/cpw/alakazam/logic/conf"
	"gitlab.com/jetfueltw/cpw/micro/config"
	"gitlab.com/jetfueltw/cpw/micro/database"
	"gitlab.com/jetfueltw/cpw/micro/http"
	"gitlab.com/jetfueltw/cpw/micro/log"
	"gitlab.com/jetfueltw/cpw/micro/redis"
)

var (
	Conf *Config
)

type Config struct {
	HTTPServer *http.Conf
	DB         *database.Conf
	Kafka      *conf.Kafka
	Redis      *redis.Conf
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
	Conf.Redis, err = redis.ReadViper(v.Sub("redis"))
	if err != nil {
		return err
	}
	Conf.DB, err = database.ReadViper(v.Sub("database"))
	if err != nil {
		return err
	}
	Conf.Kafka = &conf.Kafka{
		Topic:   v.GetString("kafka.topic"),
		Brokers: v.GetStringSlice("kafka.brokers"),
	}
	l, err := log.ReadViper(v.Sub("log"))
	if err != nil {
		return err
	}
	return log.Start(l)
}
