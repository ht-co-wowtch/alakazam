package job

import (
	"context"
	"gitlab.com/jetfueltw/cpw/alakazam/app/job/conf"
	"gitlab.com/jetfueltw/cpw/alakazam/message"
)

// Job is push job.
type Job struct {
	consume  *consume
	consumer *message.Consumer
	ctx      context.Context
	cancel   context.CancelFunc
}

// New new a push job.
func New(c *conf.Config) *Job {
	ctx, cancel := context.WithCancel(context.Background())
	j := &Job{
		consumer: message.NewConsumer(ctx, c.Kafka.Topic, c.Kafka.Group, c.Kafka.Brokers),
		consume: &consume{
			rc:    c.Room,
			ctx:   ctx,
			rooms: make(map[string]*Room),
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

func (j *Job) Run() {
	go j.consumer.Run(j.consume)
}

func (j *Job) Close() {
	j.cancel()
	j.consumer.Close()
}
