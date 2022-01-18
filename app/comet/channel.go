package comet

import (
	"sync"

	"gitlab.com/ht-co/micro/log"
	"gitlab.com/ht-co/wowtch/live/alakazam/app/comet/pb"
	"gitlab.com/ht-co/wowtch/live/alakazam/pkg/bufio"
	"go.uber.org/zap"
)

// Channel
// 用於推送消息給user，可以把這個識別user在聊天室內的地址
// 紀錄了當初連線至聊天室時所給的參數值
// 1. 身處在哪一個聊天室
// 2. user uid (user id)
// 3. user key
type Channel struct {
	// 該user進入的房間
	Room *Room

	// 讀寫異步的grpc.Proto緩型Pool
	protoRing Ring

	// 透過此管道處理Job service 送過來的資料
	signal chan *pb.Proto

	// 用於寫操作的byte
	Writer bufio.Writer

	// 用於讀操作的byte
	Reader bufio.Reader

	// 雙向鏈結串列 rlink
	Next *Channel

	// 雙向鏈結串列 llink
	Prev *Channel

	// user id
	Uid string

	// user在logic service的key
	Key string

	// 用戶名稱
	Name string

	// 用戶類型
	Type int32

	// user ip
	IP string

	// 讀寫鎖
	mutex sync.RWMutex
}

// NewChannel
// New a Channel
func NewChannel(protoSize, revBuffer int) *Channel {
	c := new(Channel)
	c.protoRing.Init(protoSize)

	// grpc接收資料的緩充量
	c.signal = make(chan *pb.Proto, revBuffer)

	return c
}

// Push
// 針對某人推送訊息
func (c *Channel) Push(p *pb.Proto) (err error) {
	// 當發送訊息速度大於消費速度則會阻塞
	// 使用select方式來避免這一塊但此時會有訊息丟失的風險存在
	// 可以提高signal buffer來避免但會耗費內存
	select {
	// 每個Channel會有自己signal接收處理的goroutine
	case c.signal <- p:
	default:
	}
	return
}

// Ready
// 等待管道接收grpc資料
func (c *Channel) Ready() *pb.Proto {
	return <-c.signal
}

// Signal
// 接收到tcp資料傳遞給處理的goroutine
func (c *Channel) Signal() {
	c.signal <- pb.ProtoReady
}

// Close
// 關閉連線flag
func (c *Channel) Close() {
	log.Info("[channel.go]Conn close", zap.String("uid", c.Uid), zap.String("name", c.Name))
	c.signal <- pb.ProtoFinish
}
