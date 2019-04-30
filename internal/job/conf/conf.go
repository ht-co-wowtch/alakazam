package conf

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/spf13/viper"
	"io/ioutil"
	"os"
	"time"

	"github.com/bilibili/discovery/naming"
)

var (
	// config path
	confPath string

	// Conf config
	Conf *Config
)

// Config is job config.
type Config struct {
	Env       *Env
	Kafka     *Kafka
	Discovery *naming.Config
	Comet     *Comet
	Room      *Room
}

// 房間消息聚合
type Room struct {
	// Signal時間內最大緩衝的訊息量，超過就推送給comet
	// 設太小等於Signal沒用，設大太每次都一定要等Signal時間到
	// 默認是20筆
	Batch int

	// 消息聚合等待多久才推送房間消息給comet，默認一秒應該是最好的優化
	// 設小會提高job通知comet的頻率，設太大房間訊息會更延遲
	Signal time.Duration

	// 消息聚合goroutine等待多久都沒收到訊息自動close
	Idle time.Duration
}

// Comet is comet config.
type Comet struct {
	// 處理訊息推送給comet的chan的Buffer
	RoutineChan int

	// 開多個goroutine併發做send grpc client
	RoutineSize int
}

// Kafka is kafka config.
type Kafka struct {
	Topic   string
	Group   string
	Brokers []string
}

// Env is env config.
type Env struct {
	Region    string
	Zone      string
	DeployEnv string
	Host      string
}


func init() {
	flag.StringVar(&confPath, "conf", "job-example.yml", "default config path")
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
		},
		Discovery: &naming.Config{
			Nodes:  []string{":7171"},
			Region: "sh",
			Zone:   "sh001",
			Env:    "dev",
			Host:   host,
		},
		Kafka: &Kafka{
			Topic:   viper.GetString("kafka.topic"),
			Group:   viper.GetString("kafka.group"),
			Brokers: viper.GetStringSlice("kafka.brokers"),
		},
		Comet: &Comet{
			RoutineChan: viper.GetInt("comet.routineChan"),
			RoutineSize: viper.GetInt("comet.routineSize"),
		},
		Room: &Room{
			Batch:  20,
			Signal: time.Second,
			Idle:   time.Duration(viper.GetInt("room.idle")) * time.Second,
		},
	}
}
