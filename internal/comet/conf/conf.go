package conf

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/spf13/viper"
	"io/ioutil"
	"time"
)

var (
	// config path
	confPath string

	// Conf config
	Conf *Config
)

// Config is comet config.
type Config struct {
	Websocket *Websocket
	TCP       *TCP
	Protocol  *Protocol
	Bucket    *Bucket
	RPCClient *RPCClient
	RPCServer *RPCServer
}

// grpc client config
type RPCClient struct {
	// client連線timeout
	Dial time.Duration
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

// TCP is tcp config.
type TCP struct {
	// tcp寫資料的緩衝區大小，該緩衝區滿到無法發送時會阻塞，此值通常設定完後系統會自行在多一倍，設定1024會變2304
	Sndbuf int

	// tcp讀取資料的緩衝區大小，該緩衝區為0時會阻塞，此值通常設定完後，系統會自行在多一倍，設定1024會變2304
	Rcvbuf int

	// 是否開啟KeepAlive
	KeepAlive bool

	// 先初始化多少個用於Reader bytes的Pool
	// 每個Pool都會有sync.Mutex，多個pool來分散鎖的競爭
	// 有效提高併發數
	Reader int

	// 每個Reader bytes Pool有多少個Buffer
	ReadBuf int

	// 每個Reader bytes Pool的Buffer能有多大的空間
	ReadBufSize int

	// 先初始化多少個用於Writer bytes的Pool
	// 每個Pool都會有sync.Mutex，多個pool來分散鎖的競爭
	// 有效提高併發數
	Writer int

	// 每個Writer bytes Pool有多少個Buffer
	WriteBuf int

	// 每個Writer bytes Pool的Buffer能有多大的空間
	WriteBufSize int
}

// Websocket is websocket config.
type Websocket struct {
	// Websocket 要監聽的port
	Bind []string
}

// Protocol is protocol config.
type Protocol struct {
	// 先初始化多少個time.Timer
	Timer int

	// 每個time.Timer一開始能接收的TimerData數量
	TimerSize int

	// 每一個連線開grpc接收資料的緩充量，當寫的速度大於讀的速度這時會阻塞，透過調大此值可以有更多緩衝避免阻塞
	SvrProto int

	// 每一個連線開異步Proto結構緩型Pool的大小，跟client透過tcp or websocket傳遞資料做消費速度有關聯
	// 由於寫的速度有可能大於讀的速度，這時會自行close此連線，透過調大此值可以有更多緩衝close
	CliProto int

	// 一開始tcp連線後等待多久沒有請求連至某房間，連線就直接close
	//
	//             -> 送auth資料 ok
	// tcp -> 等待 ->
	//             -> 超時close
	//
	HandshakeTimeout time.Duration
}

// Bucket is bucket config.
type Bucket struct {
	// 固定幾個bucket做分散
	Size int

	// 每個Bucket預先管理多少個user，不夠會自動加倍開
	Channel int

	// 每個Bucket一開始管理多少個房間，不夠會自動加倍開
	Room int

	// 每個Bucket開幾個goroutine併發做房間推送
	RoutineAmount uint64

	// 每個goroutine推送管道最大緩衝量
	RoutineSize int
}

func init() {
	flag.StringVar(&confPath, "c", "comet.yml", "default config path.")
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
	return &Config{
		RPCClient: &RPCClient{
			Dial: time.Duration(viper.GetInt("rpcClient.timeout")) * time.Second,
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
		TCP: &TCP{
			Sndbuf:       4096,
			Rcvbuf:       4096,
			KeepAlive:    false,
			Reader:       32,
			ReadBuf:      512,
			ReadBufSize:  4096,
			Writer:       32,
			WriteBuf:     512,
			WriteBufSize: 4096,
		},
		Websocket: &Websocket{
			Bind: viper.GetStringSlice("websocket.host"),
		},
		Protocol: &Protocol{
			Timer:            viper.GetInt("protocol.timer"),
			TimerSize:        viper.GetInt("protocol.timerSize"),
			CliProto:         viper.GetInt("protocol.clientProto"),
			SvrProto:         viper.GetInt("protocol.receiveProtoBuffer"),
			HandshakeTimeout: time.Second * time.Duration(viper.GetInt("protocol.handshakeTimeout")),
		},
		Bucket: &Bucket{
			Size:          viper.GetInt("bucket.size"),
			Channel:       viper.GetInt("bucket.channel"),
			Room:          viper.GetInt("bucket.room"),
			RoutineAmount: uint64(viper.GetInt("bucket.routineAmount")),
			RoutineSize:   viper.GetInt("bucket.routineSize"),
		},
	}
}
