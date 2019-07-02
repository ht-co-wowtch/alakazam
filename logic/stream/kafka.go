package stream

import (
	"github.com/gogo/protobuf/proto"
	log "github.com/golang/glog"
	"gitlab.com/jetfueltw/cpw/alakazam/protocol"
	"gitlab.com/jetfueltw/cpw/alakazam/protocol/grpc"
	"gopkg.in/Shopify/sarama.v1"
	"strconv"
)

// 房間推送，以下為條件
// 1. room id
func (d *Stream) BroadcastRoomMsg(room string, msg []byte, model grpc.PushMsg_Type) error {
	pushMsg := &grpc.PushMsg{
		Type: model,
		Room: []string{room},
		Msg:  msg,
	}
	b, err := proto.Marshal(pushMsg)
	if err != nil {
		return err
	}
	m := &sarama.ProducerMessage{
		Key:   sarama.StringEncoder(room),
		Topic: d.c.Topic,
		Value: sarama.ByteEncoder(b),
	}
	if _, _, err = d.SendMessage(m); err != nil {
		return err
	}
	return nil
}

// 多房間推送，以下為條件
func (d *Stream) BroadcastMsg(roomIds []string, msg []byte) (err error) {
	pushMsg := &grpc.PushMsg{
		Type: grpc.PushMsg_ROOM,
		Msg:  msg,
		Room: roomIds,
	}
	b, err := proto.Marshal(pushMsg)
	if err != nil {
		return
	}
	// TODO Key
	m := &sarama.ProducerMessage{
		Key:   sarama.StringEncoder(strconv.FormatInt(int64(protocol.OpRaw), 10)),
		Topic: d.c.Topic,
		Value: sarama.ByteEncoder(b),
	}
	if _, _, err = d.SendMessage(m); err != nil {
		log.Errorf("PushMsg.send(broadcast pushMsg:%v) error(%v)", pushMsg, err)
	}
	return
}
