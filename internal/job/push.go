package job

import (
	"context"
	"fmt"
	"gitlab.com/jetfueltw/cpw/alakazam/protocol"
	"gitlab.com/jetfueltw/cpw/alakazam/protocol/grpc"

	log "github.com/golang/glog"
	"gitlab.com/jetfueltw/cpw/alakazam/pkg/bytes"
)

// 訊息推送至comet server
func (j *Job) push(ctx context.Context, pushMsg *grpc.PushMsg) (err error) {
	switch pushMsg.Type {
	// 單一人推送
	case grpc.PushMsg_PUSH:
		err = j.pushKeys(pushMsg.Server, pushMsg.Keys, pushMsg.Msg)
	// 單一房間推送
	case grpc.PushMsg_ROOM:
		err = j.getRoom(pushMsg.Room).Push(pushMsg.Msg)
	// 所有房間推送
	case grpc.PushMsg_BROADCAST:
		err = j.broadcast(pushMsg.Msg, pushMsg.Speed)
	// 異常資料
	default:
		err = fmt.Errorf("no match push type: %s", pushMsg.Type)
	}
	return
}

// 單人訊息推送至comet server
func (j *Job) pushKeys(serverID string, subKeys []string, body []byte) (err error) {
	buf := bytes.NewWriterSize(len(body) + 64)
	p := &grpc.Proto{
		Op:   protocol.OpRaw,
		Body: body,
	}
	p.WriteTo(buf)
	p.Body = buf.Buffer()
	p.Op = protocol.OpRaw
	var args = grpc.PushMsgReq{
		Keys:  subKeys,
		Proto: p,
	}

	// 根據user所在的comet server id做發送
	if c, ok := j.cometServers[serverID]; ok {
		if err = c.Push(&args); err != nil {
			log.Errorf("c.Push(%v) serverID:%s error(%v)", args, serverID, err)
		}
		log.Infof("pushKey:%s comets:%d", serverID, len(j.cometServers))
	}
	return
}

// 多房間訊息推送給comet
func (j *Job) broadcast(body []byte, speed int32) (err error) {
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
	for serverID, c := range comets {
		if err = c.Broadcast(&args); err != nil {
			log.Errorf("c.Broadcast(%v) serverID:%s error(%v)", args, serverID, err)
		}
	}
	log.Infof("broadcast comets:%d", len(comets))
	return
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
	for serverID, c := range comets {
		if err = c.BroadcastRoom(&args); err != nil {
			log.Errorf("c.BroadcastRoom(%v) roomID:%s serverID:%s error(%v)", args, roomID, serverID, err)
		}
	}
	log.Infof("broadcastRoom comets:%d", len(comets))
	return
}
