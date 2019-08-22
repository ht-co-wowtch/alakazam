package job

import (
	"context"
	"fmt"
	"gitlab.com/jetfueltw/cpw/alakazam/app/comet/pb"
	"gitlab.com/jetfueltw/cpw/alakazam/app/job/conf"
	"gitlab.com/jetfueltw/cpw/micro/grpc"
	"gitlab.com/jetfueltw/cpw/micro/log"
	"go.uber.org/zap"
	"sync/atomic"
	"time"
)

// 與Comet server 建立grpc client
func newCometClient(c *grpc.Conf) (pb.CometClient, error) {
	conn, err := grpc.NewClient(c)
	if err != nil {
		return nil, err
	}
	return pb.NewCometClient(conn), nil
}

// Comet is a comet.
type Comet struct {
	// 某Comet server的 ip or name
	name string

	// Comet grpc client
	client pb.CometClient

	// 處理單一房間訊息推送給comet的chan
	roomChan []chan *pb.BroadcastRoomReq

	// 處理多房間訊息推送給comet的chan
	broadcastChan chan *pb.BroadcastReq

	// 處理踢人
	closeChan chan *pb.KeyReq

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
		roomChan:      make([]chan *pb.BroadcastRoomReq, c.RoutineSize),
		broadcastChan: make(chan *pb.BroadcastReq, c.RoutineSize),
		closeChan:     make(chan *pb.KeyReq, 100),
		routineSize:   uint64(c.RoutineSize),
	}

	// 跟Comet servers建立grpc client
	var err error
	if cmt.client, err = newCometClient(c.Comet); err != nil {
		return nil, err
	}
	cmt.ctx, cmt.cancel = context.WithCancel(context.Background())

	// 開多個goroutine併發做send grpc client
	for i := 0; i < c.RoutineSize; i++ {
		cmt.roomChan[i] = make(chan *pb.BroadcastRoomReq, c.RoutineChan)
		go cmt.process(cmt.roomChan[i], cmt.broadcastChan, cmt.closeChan)
	}
	return cmt, nil
}

// 房間訊息推送需要交由某個處理推送邏輯的goroutine
// Comet自己會預先開好多個goroutine，每個goroutine內都有一把
// 用於房間訊息推送chan，用原子鎖遞增%goroutine總數量來輪替
// 由哪一個goroutine，也就是平均分配推送的量給各goroutine
func (c *Comet) BroadcastRoom(arg *pb.BroadcastRoomReq) {
	idx := atomic.AddUint64(&c.roomChanNum, 1) % c.routineSize
	c.roomChan[idx] <- arg
}

// 多個房間推送
func (c *Comet) Broadcast(arg *pb.BroadcastReq) {
	c.broadcastChan <- arg
}

func (c *Comet) Kick(arg *pb.KeyReq) {
	c.closeChan <- arg
}

// 處理訊息推送給comet server
func (c *Comet) process(roomChan chan *pb.BroadcastRoomReq, broadcastChan chan *pb.BroadcastReq, closeChan chan *pb.KeyReq) {
	for {
		select {
		// 多個房間推送
		case broadcastArg := <-broadcastChan:
			_, err := c.client.Broadcast(context.Background(), &pb.BroadcastReq{
				Proto: broadcastArg.Proto,
				Speed: broadcastArg.Speed,
			})
			if err != nil {
				log.Error("grpc client push broadcast", zap.Error(err), zap.Any("arg", broadcastArg))
			}
		// 單一房間推送
		case roomArg := <-roomChan:
			_, err := c.client.BroadcastRoom(context.Background(), &pb.BroadcastRoomReq{
				RoomID: roomArg.RoomID,
				Proto:  roomArg.Proto,
			})
			if err != nil {
				log.Error("grpc client push room", zap.Error(err), zap.Any("arg", roomArg))
			}
		case keysArg := <-closeChan:
			_, err := c.client.Kick(context.Background(), keysArg)
			if err != nil {
				log.Error("grpc client kick", zap.Error(err), zap.Any("arg", keysArg))
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
	return err
}
