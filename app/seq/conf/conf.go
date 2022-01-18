package conf

import (
	"gitlab.com/ht-co/micro/config"
	"gitlab.com/ht-co/micro/database"
	"gitlab.com/ht-co/micro/grpc"
	"gitlab.com/ht-co/micro/log"
	"gitlab.com/ht-co/micro/redis"
)

var (
	Conf *Config
)

type Config struct {
	RPCServer   *grpc.Conf
	DB          *database.Conf
	Redis       *redis.Conf
	MetricsAddr string
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
	Conf.MetricsAddr = v.GetString("metrics.addr")
	Conf.RPCServer, err = grpc.ReadViper(v.Sub("grpc.server"))
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
	l, err := log.ReadViper(v.Sub("log"))
	if err != nil {
		return err
	}
	return log.Start(l)
}
