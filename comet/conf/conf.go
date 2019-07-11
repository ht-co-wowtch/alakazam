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

// Config is comet config.
type Config struct {
	Websocket *Websocket
	TCP       *TCP
	Protocol  *Protocol
	Bucket    *Bucket
	RPCClient *grpc.Conf
	RPCServer *grpc.Conf
}

// tcp config
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

// websocket config
type Websocket struct {
	// Websocket 要監聽的port
	Addr string
}

// protocol config
type Protocol struct {
	// 先初始化多少個time.Timer
	Timer int

	// 每個time.Timer一開始能接收的TimerData數量
	TimerSize int

	// 每一個連線開grpc接收資料的緩充量，當寫的速度大於讀的速度這時會阻塞，透過調大此值可以有更多緩衝避免阻塞
	RevBuffer int

	// 每一個連線開異步Proto結構緩型Pool的大小，跟client透過tcp or websocket傳遞資料做消費速度有關聯
	// 由於寫的速度有可能大於讀的速度，這時會自行close此連線，透過調大此值可以有更多緩衝close
	ProtoSize int

	// 一開始tcp連線後等待多久沒有請求連至某房間，連線就直接close
	//
	//             -> 送auth資料 ok
	// tcp -> 等待 ->
	//             -> 超時close
	//
	HandshakeTimeout time.Duration
}

// 紀錄各用戶Channel與Room來併發做房間推送
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
	config.SetEnvReplace(true)
	config.SetEnvPrefix("alakazam")
}

func Read(path string) error {
	v, err := config.Read(path)
	if err != nil {
		return err
	}
	Conf = new(Config)
	Conf.RPCClient, _ = grpc.ReadViper(v.Sub("grpc.client"))
	Conf.RPCServer, _ = grpc.ReadViper(v.Sub("grpc.server"))
	Conf.TCP = &TCP{
		Sndbuf:       4096,
		Rcvbuf:       4096,
		KeepAlive:    false,
		Reader:       32,
		ReadBuf:      512,
		ReadBufSize:  4096,
		Writer:       32,
		WriteBuf:     512,
		WriteBufSize: 4096,
	}
	Conf.Websocket = &Websocket{
		Addr: v.GetString("websocket.addr"),
	}

	p := v.Sub("protocol")
	ht, err := time.ParseDuration(p.GetString("handshakeTimeout"))
	if err != nil {
		return err
	}
	Conf.Protocol = &Protocol{
		Timer:            p.GetInt("timer"),
		TimerSize:        p.GetInt("timerSize"),
		ProtoSize:        p.GetInt("clientProto"),
		RevBuffer:        p.GetInt("receiveProtoBuffer"),
		HandshakeTimeout: ht,
	}
	b := v.Sub("bucket")
	Conf.Bucket = &Bucket{
		Size:          b.GetInt("size"),
		Channel:       b.GetInt("channel"),
		Room:          b.GetInt("room"),
		RoutineAmount: uint64(b.GetInt("routineAmount")),
		RoutineSize:   b.GetInt("routineSize"),
	}
	l, err := log.ReadViper(v.Sub("log"))
	if err != nil {
		return err
	}
	return log.Start(l)
}
