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

// 單一推送，以下為條件
// 1. server name
// 2. user key
// 3. operation
func (d *Dao) PushMsg(c context.Context, server string, ids []string, msg []byte) (err error) {
	pushMsg := &grpc.PushMsg{
		Type:   grpc.PushMsg_PUSH,
		Server: server,
		Ids:    ids,
		Msg:    msg,
	}
	b, err := proto.Marshal(pushMsg)
	if err != nil {
		return
	}

	// 推送給kafka
	m := &sarama.ProducerMessage{
		Key:   sarama.StringEncoder(ids[0]),
		Topic: d.c.Kafka.Topic,
		Value: sarama.ByteEncoder(b),
	}
	if _, _, err = d.kafkaPub.SendMessage(m); err != nil {
		log.Errorf("PushMsg.send(push pushMsg:%v) error(%v)", pushMsg, err)
	}
	return
}

// 房間推送，以下為條件
// 1. room id
// 2. operation
func (d *Dao) BroadcastRoomMsg(c context.Context, room string, msg []byte) (err error) {
	pushMsg := &grpc.PushMsg{
		Type: grpc.PushMsg_ROOM,
		Room: room,
		Msg:  msg,
	}
	b, err := proto.Marshal(pushMsg)
	if err != nil {
		return
	}

	// 推送給kafka
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

// 所有房間推送，以下為條件
// 1. operation
func (d *Dao) BroadcastMsg(c context.Context, speed int32, msg []byte) (err error) {
	pushMsg := &grpc.PushMsg{
		Type:  grpc.PushMsg_BROADCAST,
		Speed: speed,
		Msg:   msg,
	}
	b, err := proto.Marshal(pushMsg)
	if err != nil {
		return
	}
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
