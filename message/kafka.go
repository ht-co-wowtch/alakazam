package message

import (
	kafka "github.com/Shopify/sarama"
	"github.com/gogo/protobuf/proto"
	cometpb "gitlab.com/jetfueltw/cpw/alakazam/comet/pb"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/conf"
	logicpb "gitlab.com/jetfueltw/cpw/alakazam/logic/pb"
	"strconv"
)

type Producer struct {
	c *conf.Kafka

	kafka.SyncProducer
}

func NewProducer(c *conf.Kafka) *Producer {
	kc := kafka.NewConfig()
	kc.Version = kafka.V2_3_0_0
	kc.Producer.Return.Successes = true
	pub, err := kafka.NewSyncProducer(c.Brokers, kc)
	if err != nil {
		panic(err)
	}
	return &Producer{c: c, SyncProducer: pub}
}

// 房間推送，以下為條件
// 1. room id
func (d *Producer) BroadcastRoom(room string, msg []byte, model logicpb.PushMsg_Type) error {
	pushMsg := &logicpb.PushMsg{
		Type: model,
		Room: []string{room},
		Msg:  msg,
	}
	b, err := proto.Marshal(pushMsg)
	if err != nil {
		return err
	}
	m := &kafka.ProducerMessage{
		Key:   kafka.StringEncoder(room),
		Topic: d.c.Topic,
		Value: kafka.ByteEncoder(b),
	}
	if _, _, err = d.SendMessage(m); err != nil {
		return err
	}
	return nil
}

// 多房間推送，以下為條件
func (d *Producer) Broadcast(roomIds []string, msg []byte, model logicpb.PushMsg_Type) (int32, int64, error) {
	pushMsg := &logicpb.PushMsg{
		Type: model,
		Msg:  msg,
		Room: roomIds,
	}
	b, err := proto.Marshal(pushMsg)
	if err != nil {
		return 0, 0, err
	}
	// TODO Key
	m := &kafka.ProducerMessage{
		Key:   kafka.StringEncoder(strconv.FormatInt(int64(cometpb.OpRaw), 10)),
		Topic: d.c.Topic,
		Value: kafka.ByteEncoder(b),
	}
	partition, offset, err := d.SendMessage(m)
	if err != nil {
		return 0, 0, err
	}
	return partition, offset, nil
}
