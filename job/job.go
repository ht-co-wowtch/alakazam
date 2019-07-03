package job

import (
	"context"
	cluster "github.com/bsm/sarama-cluster"
	"github.com/gogo/protobuf/proto"
	"gitlab.com/jetfueltw/cpw/alakazam/job/conf"
	"gitlab.com/jetfueltw/cpw/alakazam/protocol/grpc"
	"gitlab.com/jetfueltw/cpw/micro/log"
	"go.uber.org/zap"
	"os"
	"sync"
)

// Job is push job.
type Job struct {
	c *conf.Config

	// 接收Kafka推送
	consumer *cluster.Consumer

	// 線上有運行哪些Comet server (不同host)
	cometServers map[string]*Comet

	// 房間訊息聚合
	rooms map[string]*Room

	// 讀寫鎖
	roomsMutex sync.RWMutex
}

// New new a push job.
func New(c *conf.Config) *Job {
	j := &Job{
		c:        c,
		consumer: newKafkaSub(c.Kafka),
		rooms:    make(map[string]*Room),
	}
	host, _ := os.Hostname()

	var err error
	j.cometServers = make(map[string]*Comet, 1)
	if j.cometServers[host], err = NewComet(c.Comet); err != nil {
		panic(err)
	}
	return j
}

func newKafkaSub(c *conf.Kafka) *cluster.Consumer {
	config := cluster.NewConfig()
	config.Consumer.Return.Errors = true
	config.Group.Return.Notifications = true
	consumer, err := cluster.NewConsumer(c.Brokers, c.Group, []string{c.Topic}, config)
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
		select {
		case err := <-j.consumer.Errors():
			log.Error("kafka consumer", zap.Error(err))
		case n := <-j.consumer.Notifications():
			log.Error("kafka consumer notifications", zap.Any("data", n))
		case msg, ok := <-j.consumer.Messages():
			if !ok {
				return
			}
			j.consumer.MarkOffset(msg, "")
			// process push message
			pushMsg := new(grpc.PushMsg)
			if err := proto.Unmarshal(msg.Value, pushMsg); err != nil {
				log.Error("proto unmarshal", zap.Error(err), zap.Any("data", msg))
				continue
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
	}
}
