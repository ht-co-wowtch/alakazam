package job

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	log "github.com/golang/glog"
	comet "gitlab.com/jetfueltw/cpw/alakazam/protocol/grpc"
	"gitlab.com/jetfueltw/cpw/alakazam/job/conf"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

var (
	// 心跳包的頻率
	grpcKeepAliveTime = time.Duration(10) * time.Second

	// 心跳回覆如果超過此時間則close連線
	grpcKeepAliveTimeout = time.Duration(3) * time.Second

	// 連線失敗後等待多久才又開始嘗試練線
	grpcBackoffMaxDelay = time.Duration(3) * time.Second

	// grpc htt2 相關參數
	grpcMaxSendMsgSize = 1 << 24
	grpcMaxCallMsgSize = 1 << 24
)

const (
	// grpc options
	grpcInitialWindowSize     = 1 << 24
	grpcInitialConnWindowSize = 1 << 24
)

// 與Comet server 建立grpc client
func newCometClient(c *conf.RPCClient) (comet.CometClient, error) {
	// grpc 連線的timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(c.Timeout))
	defer cancel()
	conn, err := grpc.DialContext(ctx, c.Addr,
		[]grpc.DialOption{
			// 與server溝通不用檢查憑證之類
			grpc.WithInsecure(),
			// Http2相關參數設定
			grpc.WithInitialWindowSize(grpcInitialWindowSize),
			grpc.WithInitialConnWindowSize(grpcInitialConnWindowSize),
			grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(grpcMaxCallMsgSize)),
			grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(grpcMaxSendMsgSize)),
			// grpc嘗試連線時間
			grpc.WithBackoffMaxDelay(grpcBackoffMaxDelay),
			// 心跳機制參數
			grpc.WithKeepaliveParams(keepalive.ClientParameters{
				Time:                grpcKeepAliveTime,
				Timeout:             grpcKeepAliveTimeout,
				PermitWithoutStream: true,
			}),
		}...,
	)
	if err != nil {
		return nil, err
	}
	return comet.NewCometClient(conn), err
}

// Comet is a comet.
type Comet struct {
	// 某Comet server的 ip or name
	name string

	// Comet grpc client
	client comet.CometClient

	// 處理單一房間訊息推送給comet的chan
	roomChan []chan *comet.BroadcastRoomReq

	// 處理多房間訊息推送給comet的chan
	broadcastChan chan *comet.BroadcastReq

	// 決定併發單人訊息推送至comet的goroutine參數
	// 使用原子鎖做遞增來平均分配給goroutine
	pushChanNum uint64

	// 決定併發單一房間訊息推送至comet的goroutine參數
	// 使用原子鎖做遞增來平均分配給goroutine
	roomChanNum uint64

	// 開多少goroutine來併發訊息推送給comet
	routineSize uint64

	// 上下文，用來控制與grpc併發退出
	ctx context.Context

	// 上下文退出
	cancel context.CancelFunc
}

// new a comet
func NewComet(c *conf.Comet) (*Comet, error) {
	cmt := &Comet{
		roomChan:      make([]chan *comet.BroadcastRoomReq, c.RoutineSize),
		broadcastChan: make(chan *comet.BroadcastReq, c.RoutineSize),
		routineSize:   uint64(c.RoutineSize),
	}

	// 跟Comet servers建立grpc client
	var err error
	if cmt.client, err = newCometClient(c.RPCClient); err != nil {
		return nil, err
	}
	cmt.ctx, cmt.cancel = context.WithCancel(context.Background())

	// 開多個goroutine併發做send grpc client
	for i := 0; i < c.RoutineSize; i++ {
		cmt.roomChan[i] = make(chan *comet.BroadcastRoomReq, c.RoutineChan)
		go cmt.process(cmt.roomChan[i], cmt.broadcastChan)
	}
	return cmt, nil
}

// 房間訊息推送需要交由某個處理推送邏輯的goroutine
// Comet自己會預先開好多個goroutine，每個goroutine內都有一把
// 用於房間訊息推送chan，用原子鎖遞增%goroutine總數量來輪替
// 由哪一個goroutine，也就是平均分配推送的量給各goroutine
func (c *Comet) BroadcastRoom(arg *comet.BroadcastRoomReq) {
	idx := atomic.AddUint64(&c.roomChanNum, 1) % c.routineSize
	c.roomChan[idx] <- arg
}

// 多個房間推送
func (c *Comet) Broadcast(arg *comet.BroadcastReq) {
	c.broadcastChan <- arg
}

// 處理訊息推送給comet server
func (c *Comet) process(roomChan chan *comet.BroadcastRoomReq, broadcastChan chan *comet.BroadcastReq) {
	for {
		select {
		// 多個房間推送
		case broadcastArg := <-broadcastChan:
			_, err := c.client.Broadcast(context.Background(), &comet.BroadcastReq{
				Proto: broadcastArg.Proto,
				Speed: broadcastArg.Speed,
			})
			if err != nil {
				log.Errorf("c.client.Broadcast arg: %v error(%v)", broadcastArg, err)
			}
		// 單一房間推送
		case roomArg := <-roomChan:
			_, err := c.client.BroadcastRoom(context.Background(), &comet.BroadcastRoomReq{
				RoomID: roomArg.RoomID,
				Proto:  roomArg.Proto,
			})
			if err != nil {
				log.Errorf("c.client.BroadcastRoom arg: %v error(%v)", roomArg, err)
			}
		case <-c.ctx.Done():
			return
		}
	}
}

// 關閉其他正在執行的goroutine
func (c *Comet) Close() (err error) {
	finish := make(chan bool)
	go func() {
		for {
			n := len(c.broadcastChan)
			for _, ch := range c.roomChan {
				n += len(ch)
			}
			if n == 0 {
				finish <- true
				return
			}
			time.Sleep(time.Second)
		}
	}()
	select {
	case <-finish:
		log.Info("close comet finish")
	case <-time.After(5 * time.Second):
		err = fmt.Errorf("close comet(room:%d broadcast:%d) timeout", len(c.roomChan), len(c.broadcastChan))
	}
	c.cancel()
	return
}
