package job

import (
	"context"
	"fmt"
	cometpb "gitlab.com/jetfueltw/cpw/alakazam/app/comet/pb"
	"gitlab.com/jetfueltw/cpw/alakazam/app/job/conf"
	logicpb "gitlab.com/jetfueltw/cpw/alakazam/app/logic/pb"
	"gitlab.com/jetfueltw/cpw/alakazam/pkg/bytes"
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

// 訊息推送至comet server
func (c *consume) Push(pushMsg *logicpb.PushMsg) error {
	switch pushMsg.Type {
	// 單一/多房間推送
	case logicpb.PushMsg_ROOM:
		for _, r := range pushMsg.Room {
			if err := c.getRoom(r).Push(pushMsg.Msg, cometpb.OpRaw); err != nil {
				return err
			}
		}
	// 訊息頂置
	case logicpb.PushMsg_TOP:
		for _, r := range pushMsg.Room {
			if err := c.getRoom(r).Push(pushMsg.Msg, cometpb.OpTopRaw); err != nil {
				return err
			}
		}
	// 所有房間推送
	case logicpb.PushMsg_BROADCAST:
		c.broadcast(pushMsg.Msg, pushMsg.Speed)
	case logicpb.PushMsg_MONEY:
		if len(pushMsg.Room) == 1 && pushMsg.Room[0] != "" {
			return c.getRoom(pushMsg.Room[0]).Push(pushMsg.Msg, cometpb.OpMoney)
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
	p := &cometpb.Proto{
		Op:   cometpb.OpRaw,
		Body: body,
	}
	p.WriteTo(buf)
	p.Body = buf.Buffer()
	p.Op = cometpb.OpBatchRaw
	comets := c.servers
	speed /= int32(len(comets))
	var args = cometpb.BroadcastReq{
		Proto: p,
		Speed: speed,
	}
	for _, c := range comets {
		c.Broadcast(&args)
	}
}

// 房間訊息推送給comet
func (c *consume) broadcastRoomRawBytes(roomID string, body []byte) (err error) {
	args := cometpb.BroadcastRoomReq{
		RoomID: roomID,
		Proto: &cometpb.Proto{
			Op:   cometpb.OpBatchRaw,
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
