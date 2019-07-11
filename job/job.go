package job

import (
	"context"
	"fmt"
	"github.com/gogo/protobuf/proto"
	"gitlab.com/jetfueltw/cpw/alakazam/job/conf"
	"gitlab.com/jetfueltw/cpw/alakazam/protocol/grpc"
	"gitlab.com/jetfueltw/cpw/micro/log"
	"go.uber.org/zap"
	"github.com/Shopify/sarama"
	"net"
	"sync"
)

// Job is push job.
type Job struct {
	c *conf.Config

	// 接收Kafka推送
	consumer sarama.ConsumerGroup

	// 線上有運行哪些Comet server (不同host)
	cometServers map[string]*Comet

	// 房間訊息聚合
	rooms map[string]*Room

	// 讀寫鎖
	roomsMutex sync.RWMutex

	conf *conf.Kafka

	ctx context.Context

	cancel context.CancelFunc
}

// New new a push job.
func New(c *conf.Config) *Job {
	ctx, cancel := context.WithCancel(context.Background())
	j := &Job{
		c:        c,
		consumer: newKafkaSub(c.Kafka),
		rooms:    make(map[string]*Room),
		ctx:      ctx,
		cancel:   cancel,
		conf:     c.Kafka,
	}

	var err error
	j.cometServers = make(map[string]*Comet, 1)
	// TODO hostname 先寫死 後續需要註冊中心來sync
	if j.cometServers["hostname"], err = NewComet(c.Comet); err != nil {
		panic(err)
	}
	return j
}

func newKafkaSub(c *conf.Kafka) sarama.ConsumerGroup {
	version, err := sarama.ParseKafkaVersion("2.3.0")
	if err != nil {
		log.Panicf("Error parsing Kafka version: %v", err)
	}
	config := sarama.NewConfig()
	config.Version = version
	config.Consumer.Return.Errors = true

	consumer, err := sarama.NewConsumerGroup(c.Brokers, c.Group, config)
	if err != nil {
		panic(err)
	}
	return consumer
}

// Close close resounces.
func (j *Job) Close() error {
	if j.consumer != nil {
		return j.consumer.Close()
	}
	return nil
}

// 接收Kafka推送的資料
func (j *Job) Consume() {
	for {
		if err := j.consumer.Consume(j.ctx, []string{j.conf.Topic}, j); err != nil {
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
		fmt.Println("1111")
	}
}

func (j *Job) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

func (j *Job) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (j *Job) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	select {
	case msg := <-claim.Messages():
		session.MarkMessage(msg, "")
		pushMsg := new(grpc.PushMsg)
		if err := proto.Unmarshal(msg.Value, pushMsg); err != nil {
			log.Error("proto unmarshal", zap.Error(err), zap.Any("data", msg))
			return err
		}
		log.Info("consume",
			zap.String("topic", msg.Topic),
			zap.Int32("partition", msg.Partition),
			zap.Int64("offset", msg.Offset),
			zap.String("key", string(msg.Key)),
			zap.Any("pushMsg", pushMsg),
		)
		// 開始處理推送至comet server
		if err := j.push(context.Background(), pushMsg); err != nil {
			log.Error("push", zap.Error(err))
		}
	}
	return nil
}
