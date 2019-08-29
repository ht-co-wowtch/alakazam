package message

import (
	"context"
	"errors"
	"fmt"
	"github.com/Shopify/sarama"
	"github.com/gogo/protobuf/proto"
	"gitlab.com/jetfueltw/cpw/alakazam/app/logic/pb"
	"gitlab.com/jetfueltw/cpw/micro/log"
	"go.uber.org/zap"
	"net"
)

type Consumer struct {
	ctx     context.Context
	topic   string
	group   sarama.ConsumerGroup
	handler sarama.ConsumerGroupHandler
}

func NewConsumer(ctx context.Context, topic string, name string, brokers []string) *Consumer {
	config := sarama.NewConfig()
	config.Version = sarama.V2_3_0_0
	config.Consumer.Return.Errors = true
	group, err := sarama.NewConsumerGroup(brokers, name, config)
	if err != nil {
		panic(err)
	}
	c := &Consumer{
		ctx:   ctx,
		topic: topic,
		group: group,
	}
	go c.errorProc()
	return c
}

type ConsumerGroupHandler interface {
	Push(msg *pb.PushMsg) error
}

func (c *Consumer) Run(handler ConsumerGroupHandler) {
	c.handler = &consumer{handler}
	for {
		if err := c.group.Consume(c.ctx, []string{c.topic}, c.handler); err != nil {
			switch err.(type) {
			case *net.OpError:
				log.Error("kafka consumer", zap.Error(err))
				return
			default:
				log.Error("kafka consumer", zap.Error(err))
			}
		}
		if c.ctx.Err() != nil {
			return
		}
	}
}

func (c *Consumer) Close() {
	c.ctx.Done()
	if err := c.group.Close(); err != nil {
		log.Error(err.Error())
	}
}

func (c *Consumer) errorProc() {
	for {
		select {
		case <-c.ctx.Done():
			return
		case err := <-c.group.Errors():
			log.Error("consumer message", zap.Error(err))
		}
	}
}

type consumer struct {
	handler ConsumerGroupHandler
}

func (c *consumer) Setup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (c *consumer) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

var errMessageNotFound = errors.New("consumer group claim read message not found")

func (c *consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	select {
	case msg, ok := <-claim.Messages():
		if !ok {
			return errMessageNotFound
		}
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
