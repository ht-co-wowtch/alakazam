package job

import (
	"context"
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
	switch pushMsg.Type {
	// 單人推送
	case logicpb.PushMsg_PUSH:
		c.pushRawByte(pushMsg.Keys, pushMsg.Msg, cometpb.OpRaw)
		break

	// 單房間推送
	case logicpb.PushMsg_ROOM:
		if pushMsg.IsRaw {
			for _, r := range pushMsg.Room {
				c.getRoom(r).consume.broadcastRoomRawByte(r, pushMsg.Msg, cometpb.OpRaw)
			}
		} else {
			for _, r := range pushMsg.Room {
				if err := c.getRoom(r).Push(pushMsg.Msg, cometpb.OpRaw); err != nil {
					return err
				}
			}
		}
		break

	case logicpb.PushMsg_KICK:
		c.pushRawByte(pushMsg.Keys, pushMsg.Msg, cometpb.OpProtoFinish)
		break

	// 異常資料
	default:
		return fmt.Errorf("no match push type: %s", pushMsg.Type)
	}
	return nil
}

// 房間訊息推送給comet
func (c *consume) broadcastRoomRawByte(roomID int32, body []byte, op int32) {
	args := cometpb.BroadcastRoomReq{
		RoomID: roomID,
		Proto: &cometpb.Proto{
			Op:   op,
			Body: body,
		},
	}

	for _, c := range c.servers {
		c.BroadcastRoom(&args)
	}
}

func (c *consume) pushRawByte(keys []string, body []byte, op int32) {
	args := cometpb.KeyReq{
		Key: keys,
		Proto: &cometpb.Proto{
			Op:   op,
			Body: body,
		},
	}

	for _, c := range c.servers {
		c.Push(&args)
	}
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
