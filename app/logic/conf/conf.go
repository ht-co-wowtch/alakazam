package conf

import (
	"fmt"
	"time"

	"gitlab.com/jetfueltw/cpw/micro/client"
	"gitlab.com/jetfueltw/cpw/micro/config"
	"gitlab.com/jetfueltw/cpw/micro/database"
	"gitlab.com/jetfueltw/cpw/micro/grpc"
	"gitlab.com/jetfueltw/cpw/micro/http"
	"gitlab.com/jetfueltw/cpw/micro/log"
	"gitlab.com/jetfueltw/cpw/micro/redis"

	"github.com/spf13/viper"
)

var (
	Conf *Config
)

// 聊天室後台
type AdminConfig struct {
	Host string
	// admin url
	Uri string
	// 禁言用戶url, 最後一個f命名表示為format 格式的字串
	Bannedf string
}

type Config struct {
	Admin      *AdminConfig
	RPCServer  *grpc.Conf
	HTTPServer *http.Conf
	DB         *database.Conf
	Kafka      *Kafka
	Redis      *redis.Conf
	Nidoran    *client.Conf
	Paras *client.Conf // TODO
	Seq        *grpc.Conf
	JwtSecret  []byte
	// comet連線用戶心跳，server會清除在線紀錄
	Heartbeat   int64
	MetricsAddr string
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
	Conf.MetricsAddr = v.GetString("metrics.addr")
	Conf.HTTPServer, err = http.ReadViper(v.Sub("http"))
	if err != nil {
		return err
	}
	Conf.RPCServer, err = grpc.ReadViper(v.Sub("grpc.server"))
	if err != nil {
		return err
	}
	Conf.Nidoran, err = client.ReadViper(v.Sub("nidoran"))
	if err != nil {
		return err
	} // TODO
	Conf.Redis, err = redis.ReadViper(v.Sub("redis"))
	if err != nil {
		return err
	}
	Conf.DB, err = database.ReadViper(v.Sub("database"))
	if err != nil {
		return err
	}
	Conf.Seq, err = grpc.ReadViper(v.Sub("grpc.seq"))
	if err != nil {
		return err
	}
	Conf.JwtSecret = []byte(v.GetString("jwt.authSecret"))
	Conf.Kafka = &Kafka{
		Topic:   v.GetString("kafka.topic"),
		Brokers: v.GetStringSlice("kafka.brokers"),
	}
	Conf.Heartbeat = (time.Duration(viper.GetInt("heartbeat")) * time.Second).Nanoseconds()
	l, err := log.ReadViper(v.Sub("log"))
	if err != nil {
		return err
	}

	// http://xxx.xxx.xxx.xxx (在stage 預設會是80 port, 所以不用在上port)
	adminUrl := fmt.Sprintf("%s", v.GetString("admin.host"))
	// http://xxx.xxx.xxx.xxx:xxx/banned/:uid/room/:id (參考Admin專案的route: https://gitlab.com/jetfueltw/cpw/alakazam/-/blob/develop/app/admin/api/server.go)
	// http://alakazam-admin-service:3112 (stage example)
	adminBannedUrl := fmt.Sprintf("%s/banned/%%s/room/%%d", adminUrl)
	Conf.Admin = &AdminConfig{
		Host:    v.GetString("admin.host"),
		Uri:     adminUrl,
		Bannedf: adminBannedUrl,
	}

	return log.Start(l)
}
