package conf

import (
	"github.com/spf13/viper"
	"time"
)

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

func newRpc() *RPCServer {
	return &RPCServer{
		Network:           "tcp",
		Addr:              viper.GetString("rpcServer.host"),
		IdleTimeout:       time.Second * 60,
		MaxLifeTime:       time.Hour * 2,
		ForceCloseWait:    time.Second * 20,
		KeepAliveInterval: time.Second * 60,
		KeepAliveTimeout:  time.Second * 20,
	}
}
