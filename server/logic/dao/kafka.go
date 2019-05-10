package dao

import (
	"context"
	"gitlab.com/jetfueltw/cpw/alakazam/protocol"
	"strconv"

	"github.com/gogo/protobuf/proto"
	log "github.com/golang/glog"
	"gitlab.com/jetfueltw/cpw/alakazam/protocol/grpc"
	"gopkg.in/Shopify/sarama.v1"
)

// 房間推送，以下為條件
// 1. room id
func (d *Dao) BroadcastRoomMsg(c context.Context, room string, msg []byte) (err error) {
	pushMsg := &grpc.PushMsg{
		Type: grpc.PushMsg_ROOM,
		Room: []string{room},
		Msg:  msg,
	}
	b, err := proto.Marshal(pushMsg)
	if err != nil {
		return
	}
	m := &sarama.ProducerMessage{
		Key:   sarama.StringEncoder(room),
		Topic: d.c.Kafka.Topic,
		Value: sarama.ByteEncoder(b),
	}
	if _, _, err = d.kafkaPub.SendMessage(m); err != nil {
		log.Errorf("PushMsg.send(broadcast_room pushMsg:%v) error(%v)", pushMsg, err)
	}
	return
}

// 多房間推送，以下為條件
func (d *Dao) BroadcastMsg(c context.Context, roomIds []string, speed int32, msg []byte) (err error) {
	pushMsg := &grpc.PushMsg{
		Type:  grpc.PushMsg_ROOM,
		Speed: speed,
		Msg:   msg,
		Room:  roomIds,
	}
	b, err := proto.Marshal(pushMsg)
	if err != nil {
		return
	}
	// TODO Key
	m := &sarama.ProducerMessage{
		Key:   sarama.StringEncoder(strconv.FormatInt(int64(protocol.OpRaw), 10)),
		Topic: d.c.Kafka.Topic,
		Value: sarama.ByteEncoder(b),
	}
	if _, _, err = d.kafkaPub.SendMessage(m); err != nil {
		log.Errorf("PushMsg.send(broadcast pushMsg:%v) error(%v)", pushMsg, err)
	}
	return
}
