package message

import (
	"context"
	"github.com/Shopify/sarama"
	"gitlab.com/jetfueltw/cpw/alakazam/app/message/conf"
	"gitlab.com/jetfueltw/cpw/alakazam/message"
	"gitlab.com/jetfueltw/cpw/alakazam/models"
	"gitlab.com/jetfueltw/cpw/micro/redis"
)

type Message struct {
	consumer *message.Consumer
	mysql    *message.MysqlConsumer
}

func New(c *conf.Config) *Message {
	ctx, _ := context.WithCancel(context.Background())
	config := message.Config{
		Topic:   c.Kafka.Topic,
		Name:    c.Kafka.Group,
		Brokers: c.Kafka.Brokers,
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

func (j *Message) Run() {
	go j.consumer.Run(j.mysql)
}

func (j *Message) Close() {
	j.consumer.Close()
}
