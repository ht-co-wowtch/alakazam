package message

import (
	kafka "github.com/Shopify/sarama"
	"github.com/gogo/protobuf/proto"
	logicpb "gitlab.com/jetfueltw/cpw/alakazam/app/logic/pb"
	seqpb "gitlab.com/jetfueltw/cpw/alakazam/app/seq/api/pb"
	"gitlab.com/jetfueltw/cpw/micro/log"
	"go.uber.org/zap"
)

type Producer struct {
	topic    string
	brokers  []string
	producer kafka.SyncProducer
	seq      seqpb.SeqClient
	bs       map[int64]*seq
}

type seq struct {
	cur int64
	max int64
}

func NewProducer(brokers []string, topic string, seq seqpb.SeqClient) *Producer {
	kc := kafka.NewConfig()
	kc.Version = kafka.V2_3_0_0
	kc.Producer.Return.Successes = true
	pub, err := kafka.NewSyncProducer(brokers, kc)
	if err != nil {
		panic(err)
	}
	return &Producer{
		brokers:  brokers,
		topic:    topic,
		producer: pub,
		seq:      seq,
	}
}

func (p *Producer) Close() error {
	return p.producer.Close()
}

// 房間推送，以下為條件
// 1. room id
func (p *Producer) send(pushMsg *logicpb.PushMsg) error {
	b, err := proto.Marshal(pushMsg)
	if err != nil {
		return err
	}
	var key kafka.StringEncoder
	switch pushMsg.Type {
	case logicpb.PushMsg_ROOM:
		key = kafka.StringEncoder(pushMsg.Room[0])
	case logicpb.PushMsg_MONEY:
		key = "red_envelope"
	case logicpb.PushMsg_BROADCAST:
		key = "all"
	case logicpb.PushMsg_TOP:
		key = "top"
	}
	m := &kafka.ProducerMessage{
		Key:   key,
		Topic: p.topic,
		Value: kafka.ByteEncoder(b),
	}
	partition, offset, err := p.producer.SendMessage(m)
	if err != nil {
		log.Error("message producer send message", zap.Error(err), zap.Int32("partition", partition), zap.Int64("offset", offset))
	}
	return err
}
