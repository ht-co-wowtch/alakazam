package conf

import (
	"bytes"
	"fmt"
	"github.com/spf13/viper"
	"io/ioutil"
	"time"
)

var (
	// Conf config
	Conf *Config
)

// Config is job config.
type Config struct {
	Kafka *Kafka
	Comet *Comet
	Room  *Room
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

	// 消息聚合goroutine(每個房間一個)等待多久都沒收到訊息自動close
	Idle time.Duration
}

// grpc client config
type RPCClient struct {
	// grpc client host
	Addr string

	// client連線timeout
	Timeout time.Duration
}

// Comet is comet config.
type Comet struct {
	// 處理訊息推送goroutine的chan Buffer多少
	RoutineChan int

	// 開多個goroutine併發處理訊息做send grpc client
	RoutineSize int

	RPCClient *RPCClient
}

// kafka config
type Kafka struct {
	Topic   string
	Group   string
	Brokers []string
}

func Read(path string) (err error) {
	viper.SetConfigType("yaml")
	b, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	if err := viper.ReadConfig(bytes.NewBuffer(b)); err != nil {
		panic(err)
	} else {
		fmt.Println("Using config file:", path)
	}
	Conf = load()
	return
}

// 載入config
func load() *Config {
	return &Config{
		Kafka: &Kafka{
			Topic:   viper.GetString("kafka.topic"),
			Group:   viper.GetString("kafka.group"),
			Brokers: viper.GetStringSlice("kafka.brokers"),
		},
		Comet: &Comet{
			RoutineChan: viper.GetInt("comet.routineChan"),
			RoutineSize: viper.GetInt("comet.routineSize"),
			RPCClient: &RPCClient{
				Addr:    viper.GetString("rpcClient.host"),
				Timeout: time.Duration(viper.GetInt("rpcClient.timeout")) * time.Second,
			},
		},
		Room: &Room{
			Batch:  20,
			Signal: time.Second,
			Idle:   time.Duration(viper.GetInt("room.idle")) * time.Second,
		},
	}
}
