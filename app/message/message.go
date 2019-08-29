package message

import (
	"context"
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
	consumer := message.NewConsumer(ctx, c.Kafka.Topic, c.Kafka.Group, c.Kafka.Brokers)
	return &Message{
		consumer: consumer,
		mysql:    message.NewMysqlConsumer(models.NewORM(c.DB), redis.New(c.Redis)),
	}
}

func (j *Message) Run() {
	go j.consumer.Run(j.mysql)
}

func (j *Message) Close() {
	j.consumer.Close()
}
