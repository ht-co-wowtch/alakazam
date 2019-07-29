package stream

import (
	"github.com/Shopify/sarama"
	"github.com/gogo/protobuf/proto"
	"gitlab.com/jetfueltw/cpw/alakazam/protocol"
	"gitlab.com/jetfueltw/cpw/alakazam/protocol/grpc"
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
func (d *Stream) BroadcastMsg(roomIds []string, msg []byte) (int32, int64, error) {
	pushMsg := &grpc.PushMsg{
		Type: grpc.PushMsg_ROOM,
		Msg:  msg,
		Room: roomIds,
	}
	b, err := proto.Marshal(pushMsg)
	if err != nil {
		return 0, 0, err
	}
	// TODO Key
	m := &sarama.ProducerMessage{
		Key:   sarama.StringEncoder(strconv.FormatInt(int64(protocol.OpRaw), 10)),
		Topic: d.c.Topic,
		Value: sarama.ByteEncoder(b),
	}
	partition, offset, err := d.SendMessage(m)
	if err != nil {
		return 0, 0, err
	}
	return partition, offset, nil
}
