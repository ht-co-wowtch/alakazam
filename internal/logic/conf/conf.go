package conf

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/bilibili/discovery/naming"
	"github.com/spf13/viper"
	"io/ioutil"
	"os"
	"time"
)

var (
	// config path
	confPath string

	// Conf config
	Conf *Config
)

// Config config.
type Config struct {
	Env        *Env
	Discovery  *naming.Config
	RPCServer  *RPCServer
	HTTPServer *HTTPServer
	Kafka      *Kafka
	Redis      *Redis
	Node       *Node
	Backoff    *Backoff
	Regions    map[string][]string
}

// Env is env config.
type Env struct {
	Region    string
	Zone      string
	DeployEnv string
	Host      string
	Weight    int64
}

// Node node config.
type Node struct {
	DefaultDomain string
	HostDomain    string
	TCPPort       int
	WSPort        int
	WSSPort       int

	// 心跳週期，連線沒有在既定的週期內回應，server就close
	// Heartbeat * HeartbeatMax = 週期時間
	HeartbeatMax int
	Heartbeat    time.Duration

	RegionWeight float64
}

// Backoff backoff.
type Backoff struct {
	//
	MaxDelay int32

	//
	BaseDelay int32

	//
	Factor float32

	//
	Jitter float32
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

// Kafka .
type Kafka struct {
	// Kafka 推送與接收Topic
	Topic string

	// 節點ip
	Brokers []string
}

// RPCServer is RPC server config.
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

// HTTPServer is http server config.
type HTTPServer struct {
	// host
	Network string

	// port
	Addr string

	// 沒用到
	ReadTimeout time.Duration

	// 沒用到
	WriteTimeout time.Duration
}

func init() {
	flag.StringVar(&confPath, "conf", "logic.yml", "default config path")
}

// init config.
func Init() (err error) {
	viper.SetConfigType("yaml")
	b, err := ioutil.ReadFile(confPath)
	if err != nil {
		panic(err)
	}
	if err := viper.ReadConfig(bytes.NewBuffer(b)); err != nil {
		panic(err)
	} else {
		fmt.Println("Using config file:", confPath)
	}
	Conf = load()
	Conf.Regions = map[string][]string{
		"sh": []string{
			"上海", "江苏", "浙江", "安徽", "江西", "湖北", "重庆", "陕西", "青海", "河南", "台湾",
		},
	}
	return
}

// 載入config
func load() *Config {
	host, _ := os.Hostname()
	return &Config{
		Env: &Env{
			Region:    "sh",
			Zone:      "sh001",
			DeployEnv: "dev",
			Host:      host,
			Weight:    10,
		},
		Discovery: &naming.Config{
			Nodes:  []string{":7171"},
			Region: "sh",
			Zone:   "sh001",
			Env:    "dev",
			Host:   host,
		},
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
		Backoff: &Backoff{
			MaxDelay:  300,
			BaseDelay: 3,
			Factor:    1.8,
			Jitter:    0.3,
		},
		Node: &Node{
			DefaultDomain: "conn.goim.io",
			HostDomain:    ".goim.io",
			Heartbeat:     time.Duration(viper.GetInt("node.heartbeat")) * time.Second,
			HeartbeatMax:  viper.GetInt("node.heartbeatMax"),
			TCPPort:       3101,
			WSPort:        3102,
			WSSPort:       3103,
			RegionWeight:  1.6,
		},
	}
}
