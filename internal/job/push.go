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
	// 單一人推送
	case grpc.PushMsg_PUSH:
		j.pushKeys(pushMsg.Server, pushMsg.Keys, pushMsg.Msg)
	// 單一房間推送
	case grpc.PushMsg_ROOM:
		err = j.getRoom(pushMsg.Room).Push(pushMsg.Msg)
	// 所有房間推送
	case grpc.PushMsg_BROADCAST:
		j.broadcast(pushMsg.Msg, pushMsg.Speed)
	// 異常資料
	default:
		err = fmt.Errorf("no match push type: %s", pushMsg.Type)
	}
	return
}

// 單人訊息推送至comet server
func (j *Job) pushKeys(serverName string, keys []string, body []byte) {
	buf := bytes.NewWriterSize(len(body) + 64)
	p := &grpc.Proto{
		Op:   protocol.OpRaw,
		Body: body,
	}
	p.WriteTo(buf)
	p.Body = buf.Buffer()
	p.Op = protocol.OpRaw
	var args = grpc.PushMsgReq{
		Keys:  keys,
		Proto: p,
	}

	// 根據user所在的comet server id做發送
	if c, ok := j.cometServers[serverName]; ok {
		log.Infof("pushKey:%v server:%s", keys, serverName)
		c.Push(&args)
	}
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
	p.Op = protocol.OpRaw
	comets := j.cometServers
	speed /= int32(len(comets))
	var args = grpc.BroadcastReq{
		ProtoOp: protocol.OpRaw,
		Proto:   p,
		Speed:   speed,
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
			Op:   protocol.OpRaw,
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
