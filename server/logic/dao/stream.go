package dao

import (
	"gitlab.com/jetfueltw/cpw/alakazam/server/logic/conf"
	kafka "gopkg.in/Shopify/sarama.v1"
)

type Stream struct {
	c *conf.Kafka

	kafka.SyncProducer
}
