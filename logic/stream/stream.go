package stream

import (
	"gitlab.com/jetfueltw/cpw/alakazam/logic/conf"
	kafka "github.com/Shopify/sarama"
)

type Stream struct {
	c *conf.Kafka

	kafka.SyncProducer
}

func NewKafkaPub(c *conf.Kafka) *Stream {
	kc := kafka.NewConfig()
	kc.Producer.RequiredAcks = kafka.WaitForAll
	kc.Producer.Retry.Max = 10
	kc.Producer.Return.Successes = true
	pub, err := kafka.NewSyncProducer(c.Brokers, kc)
	if err != nil {
		panic(err)
	}
	return &Stream{c: c, SyncProducer: pub}
}
