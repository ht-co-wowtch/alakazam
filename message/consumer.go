package message

import (
	"context"
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/gogo/protobuf/proto"
	"gitlab.com/jetfueltw/cpw/alakazam/app/logic/pb"
	"gitlab.com/jetfueltw/cpw/micro/log"
	"go.uber.org/zap"
	"time"
)

type Consumer struct {
	ctx     context.Context
	topic   string
	group   sarama.ConsumerGroup
	handler sarama.ConsumerGroupHandler
	ready   chan bool
	isRun   bool
}

type Config struct {
	Topic   string
	Name    string
	Brokers []string
	Offsets struct {
		Initial int64
	}
}

func NewConsumer(ctx context.Context, conf Config) *Consumer {
	config := sarama.NewConfig()
	config.Version = sarama.V2_3_0_0

	// 參考 https://kafka.apache.org/documentation/#consumerconfigs
	// sarama 沒看到 max.poll.interval.ms ?
	// enable.auto.commit 不支持
	// connections.max.idle.ms ?

	// session.timeout.ms
	config.Consumer.Group.Session.Timeout = 10 * time.Second
	// heartbeat.interval.ms
	config.Consumer.Group.Heartbeat.Interval = 3 * time.Second
	// auto.offset.reset
	config.Consumer.Offsets.Initial = conf.Offsets.Initial
	// fetch.min.bytes
	config.Consumer.Fetch.Min = 1
	// 相隔多久自動提交Offset
	config.Consumer.Offsets.CommitInterval = time.Second
	// 是否透過chan回傳error
	config.Consumer.Return.Errors = true

	client, err := sarama.NewClient(conf.Brokers, config)
	if err != nil {
		panic(err)
	}

	group, err := sarama.NewConsumerGroupFromClient(conf.Name, client)
	if err != nil {
		panic(err)
	}

	if err := registerConsumerMetric(client, config.MetricRegistry); err != nil {
		panic(err)
	}

	c := &Consumer{
		ctx:   ctx,
		topic: conf.Topic,
		group: group,
		ready: make(chan bool),
	}
	go c.proc()
	return c
}

type ConsumerGroupHandler interface {
	Push(msg *pb.PushMsg) error
}

func (c *Consumer) Run(handler ConsumerGroupHandler) {
	c.handler = &consumer{
		handler: handler,
		ready:   c.ready,
	}

	for {
		if err := c.group.Consume(c.ctx, []string{c.topic}, c.handler); err != nil {
			if c.isRun {
				c.isRun = false
				log.Error("kafka consumer", zap.Error(err))
			}
		}
		if c.ctx.Err() != nil {
			return
		}
	}
}

func (c *Consumer) Close() {
	if err := c.group.Close(); err != nil {
		log.Error(err.Error())
	}
	c.ctx.Done()
}

func (c *Consumer) proc() {
	for {
		select {
		case <-c.ctx.Done():
			return
		case isRun := <-c.ready:
			c.isRun = isRun
			break
		case err := <-c.group.Errors():
			log.Error("consumer message", zap.Error(err))
		}
	}
}

type consumer struct {
	handler ConsumerGroupHandler
	ready   chan bool
}

func (c *consumer) Setup(session sarama.ConsumerGroupSession) error {
	log.Info("kafka consumer setup")
	return nil
}

func (c *consumer) Cleanup(session sarama.ConsumerGroupSession) error {
	log.Info("kafka consumer cleanup")
	return nil
}

//var errMessageNotFound = errors.New("consumer group claim read message not found")

func (c *consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	c.ready <- true
	log.Infof("kafka start consumer message for [%d] partition", claim.Partition())

	for msg := range claim.Messages() {
		session.MarkMessage(msg, "")
		pushMsg := new(pb.PushMsg)
		if err := proto.Unmarshal(msg.Value, pushMsg); err != nil {
			return fmt.Errorf("proto unmarshal error:[%s] data: [%s]", err.Error(), string(msg.Value))
		}
		// 開始處理推送至comet server
		if err := c.handler.Push(pushMsg); err != nil {
			return err
		}
	}
	return nil
}
