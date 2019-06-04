package job

import (
	"context"
	"fmt"
	log "github.com/golang/glog"
	"gitlab.com/jetfueltw/cpw/alakazam/pkg/bytes"
	"gitlab.com/jetfueltw/cpw/alakazam/protocol"
	"gitlab.com/jetfueltw/cpw/alakazam/protocol/grpc"
)

// 訊息推送至comet server
func (j *Job) push(ctx context.Context, pushMsg *grpc.PushMsg) (err error) {
	switch pushMsg.Type {
	// 單一/多房間推送
	case grpc.PushMsg_ROOM:
		for _, r := range pushMsg.Room {
			if err = j.getRoom(r).Push(pushMsg.Msg, protocol.OpRaw); err != nil {
				break
			}
		}
	// 所有房間推送
	case grpc.PushMsg_BROADCAST:
		j.broadcast(pushMsg.Msg, pushMsg.Speed)
	case grpc.PushMsg_MONEY:
		if len(pushMsg.Room) == 1 && pushMsg.Room[0] != "" {
			err = j.getRoom(pushMsg.Room[0]).Push(pushMsg.Msg, protocol.OpMoney)
		} else {
			err = fmt.Errorf("money Can only be pushed to a single room: %s", pushMsg.Msg)
		}
	// 異常資料
	default:
		err = fmt.Errorf("no match push type: %s", pushMsg.Type)
	}

	return err
}

// 多房間訊息推送給comet
func (j *Job) broadcast(body []byte, speed int32) {
	buf := bytes.NewWriterSize(len(body) + 64)
	p := &grpc.Proto{
		Op:   protocol.OpRaw,
		Body: body,
	}
	p.WriteTo(buf)
	p.Body = buf.Buffer()
	p.Op = protocol.OpBatchRaw
	comets := j.cometServers
	speed /= int32(len(comets))
	var args = grpc.BroadcastReq{
		Proto: p,
		Speed: speed,
	}
	for serverName, c := range comets {
		log.Infof("broadcast server:%s ", serverName)
		c.Broadcast(&args)
	}
}

// 房間訊息推送給comet
func (j *Job) broadcastRoomRawBytes(roomID string, body []byte) (err error) {
	args := grpc.BroadcastRoomReq{
		RoomID: roomID,
		Proto: &grpc.Proto{
			Op:   protocol.OpBatchRaw,
			Body: body,
		},
	}
	comets := j.cometServers
	for serverName, c := range comets {
		log.Infof("broadcastRoom server:%s ", serverName)
		c.BroadcastRoom(&args)
	}
	return
}
