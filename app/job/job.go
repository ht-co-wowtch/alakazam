package job

import (
	"context"

	"github.com/Shopify/sarama"
	"gitlab.com/jetfueltw/cpw/alakazam/app/job/conf"
	"gitlab.com/jetfueltw/cpw/alakazam/message"
)

// Job is push job.
type Job struct {
	consume *consume

	// 這是Kafka consumer wrapper
	consumer *message.Consumer
	ctx      context.Context
	cancel   context.CancelFunc
}

// New new a push job.
func New(c *conf.Config) *Job {
	ctx, cancel := context.WithCancel(context.Background())
	config := message.Config{
		Topic:   c.Kafka.Topic,
		Name:    c.Kafka.Group,
		Brokers: c.Kafka.Brokers,
		Offsets: struct {
			Initial int64
		}{
			Initial: sarama.OffsetNewest,
		},
	}
	j := &Job{
		consumer: message.NewConsumer(ctx, config),
		consume: &consume{
			rc:    c.Room,
			ctx:   ctx,
			rooms: make(map[int32]*Room),
		},
		ctx:    ctx,
		cancel: cancel,
	}

	var err error
	j.consume.servers = make(map[string]*Comet, 1)
	// TODO hostname 先寫死 後續需要註冊中心來sync
	if j.consume.servers["hostname"], err = NewComet(c.Comet); err != nil {
		panic(err)
	}
	return j
}

//After Job initialize, Job`s Run method will be execute immeditaely
func (j *Job) Run() {
	go j.consumer.Run(j.consume)
}

func (j *Job) Close() {
	j.cancel()
	j.consumer.Close()
}
