package job

import (
	"context"
	"encoding/json"
	"fmt"
	cometpb "gitlab.com/jetfueltw/cpw/alakazam/app/comet/pb"
	"gitlab.com/jetfueltw/cpw/alakazam/app/job/conf"
	logicpb "gitlab.com/jetfueltw/cpw/alakazam/app/logic/pb"
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
	rooms map[int32]*Room

	ctx context.Context
}

// 訊息推送至comet server
func (c *consume) Push(pushMsg *logicpb.PushMsg) error {
	var model int32
	switch pushMsg.Type {
	// 單一/多房間推送
	case logicpb.PushMsg_ROOM, logicpb.PushMsg_MONEY, logicpb.PushMsg_ADMIN, logicpb.PushMsg_ADMIN_TOP:
		model = cometpb.OpRaw
	case logicpb.PushMsg_Close:
		return c.kick(pushMsg)
	case logicpb.PushMsg_CLOSE_TOP:
		model = cometpb.OpCloseTopMessage
	// 異常資料
	default:
		return fmt.Errorf("no match push type: %s", pushMsg.Type)
	}
	for _, r := range pushMsg.Room {
		if err := c.getRoom(r).Push(pushMsg.Msg, model); err != nil {
			return err
		}
	}
	return nil
}

// 房間訊息推送給comet
func (c *consume) broadcastRoomRawBytes(roomID int32, body []byte) (err error) {
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
func (c *consume) getRoom(roomID int32) *Room {
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
		log.Info("new a room active", zap.Int32("id", roomID), zap.Int("active", len(c.rooms)))
	}
	return room
}

// 移除房間訊息聚合
func (c *consume) delRoom(roomID int32) {
	c.roomsMutex.Lock()
	delete(c.rooms, roomID)
	c.roomsMutex.Unlock()
}

func (c *consume) kick(pushMsg *logicpb.PushMsg) error {
	msg := struct {
		Message string `json:"message"`
	}{
		Message: pushMsg.Message,
	}
	b, _ := json.Marshal(msg)
	for _, c := range c.servers {
		c.Kick(&cometpb.KeyReq{
			Proto: &cometpb.Proto{
				Op:   cometpb.OpProtoFinish,
				Body: b,
			},
			Key: pushMsg.Keys,
		})
	}
	return nil
}
