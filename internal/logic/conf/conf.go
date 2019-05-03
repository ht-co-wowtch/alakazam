package conf

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/spf13/viper"
)

var (
	// config path
	confPath string

	// Conf config
	Conf *Config
)

type Config struct {
	RPCServer  *RPCServer
	HTTPServer *HTTPServer
	Kafka      *Kafka
	Redis      *Redis

	// comet連線用戶心跳，server會清除在線紀錄
	Heartbeat int64
}

// Redis
type Redis struct {
	// host
	Network string

	// port
	Addr string

	// pool內最大連線總數
	Active int

	// 最大保留的閒置連線數
	Idle int

	// 建立連線超時多久後放棄
	DialTimeout time.Duration

	// read多久沒回覆則放棄
	ReadTimeout time.Duration

	// write多久沒回覆則放棄
	WriteTimeout time.Duration

	// 空閒連線多久沒做事就close
	IdleTimeout time.Duration

	// redis過期時間
	Expire time.Duration
}

// Kafka
type Kafka struct {
	// Kafka 推送與接收Topic
	Topic string

	// 節點ip
	Brokers []string
}

// grpc server config
type RPCServer struct {
	// host
	Network string

	// port
	Addr string

	// 當連線閒置多久後發送一個`GOAWAY` Framer 封包告知Client說太久沒活動
	//至於Client收到`GOAWAY`後要做什麼目前要自己實現stream，server只是做通知而已，grpc server默認沒開啟此功能
	IdleTimeout time.Duration

	// 任何連線只要連線超過某時間就會強制被close，但是在close之前會先發送`GOAWAY`Framer 封包告知Client
	MaxLifeTime time.Duration

	// MaxConnectionAge要關閉之前等待的時間
	ForceCloseWait time.Duration

	// keepalive頻率(心跳週期)
	KeepAliveInterval time.Duration

	// 每次做keepalive完後等待多少秒如果server沒有回應則將此連線close掉
	KeepAliveTimeout time.Duration
}

// http server config
type HTTPServer struct {
	// host
	Network string

	// port
	Addr string

	// 沒用到
	ReadTimeout time.Duration

	// 沒用到
	WriteTimeout time.Duration

	//(Debug)開發模式
	IsStage bool
}

func init() {
	flag.StringVar(&confPath, "c", "logic.yml", "default config path")
}

// init config.
func Init() (error) {
	return Read(confPath)
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
		RPCServer: &RPCServer{
			Network:           "tcp",
			Addr:              viper.GetString("rpcServer.host"),
			IdleTimeout:       time.Second * 60,
			MaxLifeTime:       time.Hour * 2,
			ForceCloseWait:    time.Second * 20,
			KeepAliveInterval: time.Second * 60,
			KeepAliveTimeout:  time.Second * 20,
		},
		HTTPServer: &HTTPServer{
			Network:      "tcp",
			Addr:         viper.GetString("httpServer.host"),
			ReadTimeout:  time.Duration(viper.GetInt("httpServer.readTimeout")) * time.Second,
			WriteTimeout: time.Duration(viper.GetInt("httpServer.writeTimeout")) * time.Second,
			IsStage:      viper.GetBool("httpServer.isStage"),
		},
		Redis: &Redis{
			Network:      "tcp",
			Addr:         viper.GetString("redis.host"),
			Active:       viper.GetInt("redis.active"),
			Idle:         viper.GetInt("redis.idle"),
			DialTimeout:  time.Duration(viper.GetInt("redis.dialTimeout")) * time.Second,
			ReadTimeout:  time.Duration(viper.GetInt("redis.readTimeout")) * time.Second,
			WriteTimeout: time.Duration(viper.GetInt("redis.writeTimeout")) * time.Second,
			IdleTimeout:  time.Duration(viper.GetInt("redis.idleTimeout")) * time.Second,
			Expire:       time.Duration(viper.GetInt("redis.expire")) * time.Second,
		},
		Kafka: &Kafka{
			Topic:   viper.GetString("kafka.topic"),
			Brokers: viper.GetStringSlice("kafka.brokers"),
		},
		Heartbeat: (time.Duration(viper.GetInt("heartbeat")) * time.Second).Nanoseconds(),
	}
}
