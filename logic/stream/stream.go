package stream

import (
	kafka "github.com/Shopify/sarama"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/conf"
)

type Stream struct {
	c *conf.Kafka

	kafka.SyncProducer
}

func NewKafkaPub(c *conf.Kafka) *Stream {
	kc := kafka.NewConfig()
	kc.Version = kafka.V2_3_0_0
	kc.Producer.RequiredAcks = kafka.WaitForAll
	kc.Producer.Retry.Max = 10
	kc.Producer.Return.Successes = true
	pub, err := kafka.NewSyncProducer(c.Brokers, kc)
	if err != nil {
		panic(err)
	}
	return &Stream{c: c, SyncProducer: pub}
}
