package message

import (
	"context"

	"github.com/Shopify/sarama"
	"gitlab.com/ht-co/wowtch/live/alakazam/app/message/conf"
	"gitlab.com/ht-co/wowtch/live/alakazam/message"
	"gitlab.com/ht-co/wowtch/live/alakazam/models"
	"gitlab.com/ht-co/cpw/micro/redis"
	// _ "net/http/httprof"
	// _ "runtime/pprof"
)

type Message struct {
	consumer *message.Consumer
	mysql    *message.MysqlConsumer
}

func New(c *conf.Config) *Message {
	ctx, _ := context.WithCancel(context.Background())

	config := message.Config{
		Name:    c.Kafka.Group,
		Brokers: c.Kafka.Brokers,
		Topic:   c.Kafka.Topic,
		Offsets: struct {
			Initial int64
		}{
			Initial: sarama.OffsetOldest,
		},
	}

	consumer := message.NewConsumer(ctx, config)
	return &Message{
		consumer: consumer,
		mysql:    message.NewMysqlConsumer(ctx, models.NewORM(c.DB), redis.New(c.Redis)),
	}
}

func (j *Message) Close() {
	j.consumer.Close()
}

func (j *Message) Run() {

	go j.consumer.Run(j.mysql)
}
