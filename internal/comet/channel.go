package comet

import (
	"sync"

	"gitlab.com/jetfueltw/cpw/alakazam/pkg/bufio"
	"gitlab.com/jetfueltw/cpw/alakazam/protocol/grpc"
)

// 用於推送消息給user，可以把這個識別user在聊天室內的地址
// 紀錄了當初連線至聊天室時所給的參數值
// 1. 身處在哪一個聊天室
// 2. user mid (user id)
// 3. user key
type Channel struct {
	// 該user進入的房間
	Room *Room

	// 讀寫異步的grpc.Proto緩型Pool
	protoRing Ring

	// 透過此管道處理Job service 送過來的資料
	signal chan *grpc.Proto

	// 用於寫操作的byte
	Writer bufio.Writer

	// 用於讀操作的byte
	Reader bufio.Reader

	// 雙向鏈結串列 rlink
	Next *Channel

	// 雙向鏈結串列 llink
	Prev *Channel

	// user id
	Mid int64

	// user在logic service的key
	Key string

	// 用戶名稱
	Name string

	// user ip
	IP string

	// 讀寫鎖
	mutex sync.RWMutex
}

// new a channel.
func NewChannel(protoSize, revBuffer int) *Channel {
	c := new(Channel)
	c.protoRing.Init(protoSize)

	// grpc接收資料的緩充量
	c.signal = make(chan *grpc.Proto, revBuffer)
	return c
}

// 針對某人推送訊息
func (c *Channel) Push(p *grpc.Proto) (err error) {
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

// 等待管道接收grpc資料
func (c *Channel) Ready() *grpc.Proto {
	return <-c.signal
}

// 接收到tcp資料傳遞給處理的goroutine
func (c *Channel) Signal() {
	c.signal <- grpc.ProtoReady
}

// 關閉連線flag
func (c *Channel) Close() {
	c.signal <- grpc.ProtoFinish
}
