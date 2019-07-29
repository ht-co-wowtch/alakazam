package job

import (
	"context"
	"errors"
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/gogo/protobuf/proto"
	"gitlab.com/jetfueltw/cpw/alakazam/job/conf"
	"gitlab.com/jetfueltw/cpw/alakazam/pkg/bytes"
	"gitlab.com/jetfueltw/cpw/alakazam/protocol"
	"gitlab.com/jetfueltw/cpw/alakazam/protocol/grpc"
	"gitlab.com/jetfueltw/cpw/micro/log"
	"go.uber.org/zap"
	"sync"
)

// 處理訊息結構
type consume struct {
	rc *conf.Room

	// 線上有運行哪些Comet server (不同host)
	servers map[string]*Comet

	// 讀寫鎖
	roomsMutex sync.RWMutex

	// 房間訊息聚合
	rooms map[string]*Room

	ctx context.Context
}

func (c *consume) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

func (c *consume) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

var errMessageNotFound = errors.New("consumer group claim read message not found")

func (c *consume) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	select {
	case msg, ok := <-claim.Messages():
		if !ok {
			return errMessageNotFound
		}
		session.MarkMessage(msg, "")
		pushMsg := new(grpc.PushMsg)
		if err := proto.Unmarshal(msg.Value, pushMsg); err != nil {
			log.Error("proto unmarshal", zap.Error(err), zap.Any("data", msg))
			return err
		}
		log.Info("consume",
			zap.String("topic", msg.Topic),
			zap.Int32("partition", msg.Partition),
			zap.Int64("offset", msg.Offset),
			zap.String("key", string(msg.Key)),
			zap.Any("pushMsg", pushMsg),
		)
		// 開始處理推送至comet server
		if err := c.push(pushMsg); err != nil {
			log.Error("push", zap.Error(err))
		}
	}
	return nil
}

// 訊息推送至comet server
func (c *consume) push(pushMsg *grpc.PushMsg) error {
	switch pushMsg.Type {
	// 單一/多房間推送
	case grpc.PushMsg_ROOM:
		for _, r := range pushMsg.Room {
			if err := c.getRoom(r).Push(pushMsg.Msg, protocol.OpRaw); err != nil {
				return err
			}
		}
	// 訊息頂置
	case grpc.PushMsg_TOP:
		for _, r := range pushMsg.Room {
			if err := c.getRoom(r).Push(pushMsg.Msg, protocol.OpTopRaw); err != nil {
				return err
			}
		}
	// 所有房間推送
	case grpc.PushMsg_BROADCAST:
		c.broadcast(pushMsg.Msg, pushMsg.Speed)
	case grpc.PushMsg_MONEY:
		if len(pushMsg.Room) == 1 && pushMsg.Room[0] != "" {
			return c.getRoom(pushMsg.Room[0]).Push(pushMsg.Msg, protocol.OpMoney)
		} else {
			return fmt.Errorf("money can only be pushed to a single room: %s", pushMsg.Msg)
		}
	// 異常資料
	default:
		return fmt.Errorf("no match push type: %s", pushMsg.Type)
	}
	return nil
}

// 多房間訊息推送給comet
func (c *consume) broadcast(body []byte, speed int32) {
	buf := bytes.NewWriterSize(len(body) + 64)
	p := &grpc.Proto{
		Op:   protocol.OpRaw,
		Body: body,
	}
	p.WriteTo(buf)
	p.Body = buf.Buffer()
	p.Op = protocol.OpBatchRaw
	comets := c.servers
	speed /= int32(len(comets))
	var args = grpc.BroadcastReq{
		Proto: p,
		Speed: speed,
	}
	for _, c := range comets {
		c.Broadcast(&args)
	}
}

// 房間訊息推送給comet
func (c *consume) broadcastRoomRawBytes(roomID string, body []byte) (err error) {
	args := grpc.BroadcastRoomReq{
		RoomID: roomID,
		Proto: &grpc.Proto{
			Op:   protocol.OpBatchRaw,
			Body: body,
		},
	}
	comets := c.servers
	for _, c := range comets {
		c.BroadcastRoom(&args)
	}
	return
}

// 根據room id取Roomd用做房間訊息聚合
func (c *consume) getRoom(roomID string) *Room {
	c.roomsMutex.RLock()
	room, ok := c.rooms[roomID]
	c.roomsMutex.RUnlock()
	if !ok {
		c.roomsMutex.Lock()
		if room, ok = c.rooms[roomID]; !ok {
			room = NewRoom(c, roomID)
			c.rooms[roomID] = room
		}
		c.roomsMutex.Unlock()
		log.Info("new a room active", zap.String("id", roomID), zap.Int("active", len(c.rooms)))
	}
	return room
}

// 移除房間訊息聚合
func (c *consume) delRoom(roomID string) {
	c.roomsMutex.Lock()
	delete(c.rooms, roomID)
	c.roomsMutex.Unlock()
}
