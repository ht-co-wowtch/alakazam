package message

import (
	"context"
	"gitlab.com/jetfueltw/cpw/alakazam/app/message/conf"
	"gitlab.com/jetfueltw/cpw/alakazam/message"
	"gitlab.com/jetfueltw/cpw/micro/database"
	"time"
)

type Message struct {
	consumer *message.Consumer
	mysql    *message.MysqlConsumer
}

func New(c *conf.Config) *Message {
	ctx, _ := context.WithCancel(context.Background())
	consumer := message.NewConsumer(ctx, c.Kafka.Topic, c.Kafka.Group, c.Kafka.Brokers)
	d, _ := database.NewORM(c.DB)
	d.SetTZDatabase(time.Local)
	return &Message{
		consumer: consumer,
		mysql:    message.NewMysqlConsumer(d),
	}
}

func (j *Message) Run() {
	go j.consumer.Run(j.mysql)
}

func (j *Message) Close() {
	j.consumer.Close()
}
