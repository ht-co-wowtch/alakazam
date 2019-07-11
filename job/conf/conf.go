package conf

import (
	"gitlab.com/jetfueltw/cpw/micro/config"
	"gitlab.com/jetfueltw/cpw/micro/grpc"
	"gitlab.com/jetfueltw/cpw/micro/log"
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

// Comet is comet config.
type Comet struct {
	// 處理訊息推送goroutine的chan Buffer多少
	RoutineChan int

	// 開多個goroutine併發處理訊息做send grpc client
	RoutineSize int

	RPCClient *grpc.Conf
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
	client, _ := grpc.ReadViper(v.Sub("grpc.client"))
	co := v.Sub("comet")
	Conf.Comet = &Comet{
		RoutineChan: co.GetInt("routineChan"),
		RoutineSize: co.GetInt("routineSize"),
		RPCClient:   client,
	}

	k := v.Sub("kafka")
	Conf.Kafka = &Kafka{
		Topic:   k.GetString("topic"),
		Group:   k.GetString("group"),
		Brokers: k.GetStringSlice("brokers"),
	}

	idle, err := time.ParseDuration(v.GetString("room.idle"))
	if err != nil {
		return err
	}
	Conf.Room = &Room{
		Batch:  20,
		Signal: time.Second,
		Idle:   idle,
	}
	l, err := log.ReadViper(v.Sub("log"))
	if err != nil {
		return err
	}
	return log.Start(l)
}
