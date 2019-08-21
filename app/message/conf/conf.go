package conf

import (
	"gitlab.com/jetfueltw/cpw/micro/config"
	"gitlab.com/jetfueltw/cpw/micro/database"
	"gitlab.com/jetfueltw/cpw/micro/log"
)

var (
	// Conf config
	Conf *Config
)

// Config is job config.
type Config struct {
	DB    *database.Conf
	Kafka *Kafka
}

// kafka config
type Kafka struct {
	Topic   string
	Group   string
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
	Conf.DB, err = database.ReadViper(v.Sub("database"))
	if err != nil {
		return err
	}

	k := v.Sub("kafka")
	Conf.Kafka = &Kafka{
		Topic:   k.GetString("topic"),
		Group:   k.GetString("group"),
		Brokers: k.GetStringSlice("brokers"),
	}

	l, err := log.ReadViper(v.Sub("log"))
	if err != nil {
		return err
	}
	return log.Start(l)
}
