package job

import (
	"context"
	"github.com/Shopify/sarama"
	"gitlab.com/jetfueltw/cpw/alakazam/job/conf"
	"gitlab.com/jetfueltw/cpw/micro/log"
	"go.uber.org/zap"
	"net"
)

// Job is push job.
type Job struct {
	c *conf.Kafka

	// 接收Kafka推送
	kafka sarama.ConsumerGroup

	// 處理訊息
	consume *consume

	ctx context.Context

	cancel context.CancelFunc
}

// New new a push job.
func New(c *conf.Config) *Job {
	ctx, cancel := context.WithCancel(context.Background())
	j := &Job{
		c:     c.Kafka,
		kafka: newKafkaSub(c.Kafka),
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

func newKafkaSub(c *conf.Kafka) sarama.ConsumerGroup {
	config := sarama.NewConfig()
	config.Version = sarama.V2_3_0_0
	config.Consumer.Return.Errors = true
	consumer, err := sarama.NewConsumerGroup(c.Brokers, c.Group, config)
	if err != nil {
		panic(err)
	}
	return consumer
}

// Close close resounces.
func (j *Job) Close() error {
	if j.kafka != nil {
		j.cancel()
		return j.kafka.Close()
	}
	return nil
}

// 接收Kafka推送的資料
func (j *Job) Consume() {
	for {
		if err := j.kafka.Consume(j.ctx, []string{j.c.Topic}, j.consume); err != nil {
			switch err.(type) {
			case *net.OpError:
				log.Panic("kafka consumer", zap.Error(err))
				return
			default:
				log.Error("kafka consumer", zap.Error(err))
			}
		}
		if j.ctx.Err() != nil {
			return
		}
	}
}
