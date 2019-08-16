package message

import (
	kafka "github.com/Shopify/sarama"
	"github.com/gogo/protobuf/proto"
	cometpb "gitlab.com/jetfueltw/cpw/alakazam/app/comet/pb"
	logicpb "gitlab.com/jetfueltw/cpw/alakazam/app/logic/pb"
	"gitlab.com/jetfueltw/cpw/micro/log"
	"go.uber.org/zap"
	"strconv"
	"time"
)

type Producer struct {
	topic    string
	brokers  []string
	producer kafka.SyncProducer
	cron     *cron
	stop     chan struct{}
}

func NewProducer(brokers []string, topic string) *Producer {
	kc := kafka.NewConfig()
	kc.Version = kafka.V2_3_0_0
	kc.Producer.Return.Successes = true
	pub, err := kafka.NewSyncProducer(brokers, kc)
	if err != nil {
		panic(err)
	}
	return &Producer{
		brokers:  brokers,
		topic:    topic,
		producer: pub,
	}
}

func (p *Producer) StartDelay() {
	cron := newCron(time.Second * 5)
	cron.start()
	p.cron = cron
	p.stop = make(chan struct{})

	go func() {
		for {
			select {
			case <-p.stop:
				return
			case m := <-p.cron.Message():
				for _, v := range m {
					var err error
					if v.redEnvelope.Id == "" {
						err = p.SendRedEnvelope(v.room[0], v.message, v.redEnvelope)
					} else {
						err = p.Send(v.room[0], v.message)
					}
					if err != nil {
						log.Error("delay send message", zap.Int64("id", v.message.Id))
					} else {
						log.Info("delay send message", zap.Int64("id", v.message.Id))
					}
				}
			}
		}
	}()
}

func (p *Producer) Close() error {
	err := p.producer.Close()
	if p.cron != nil {
		p.cron.close()
		p.stop <- struct{}{}
	}
	return err
}

// 房間推送，以下為條件
// 1. room id
func (d *Producer) BroadcastRoom(room string, msg []byte, model logicpb.PushMsg_Type) error {
	pushMsg := &logicpb.PushMsg{
		Type: model,
		Room: []string{room},
		Msg:  msg,
	}
	b, err := proto.Marshal(pushMsg)
	if err != nil {
		return err
	}
	m := &kafka.ProducerMessage{
		Key:   kafka.StringEncoder(room),
		Topic: d.topic,
		Value: kafka.ByteEncoder(b),
	}
	if _, _, err = d.producer.SendMessage(m); err != nil {
		return err
	}
	return nil
}

// 多房間推送，以下為條件
func (d *Producer) Broadcast(roomIds []string, msg []byte, model logicpb.PushMsg_Type) (int32, int64, error) {
	pushMsg := &logicpb.PushMsg{
		Type: model,
		Msg:  msg,
		Room: roomIds,
	}
	b, err := proto.Marshal(pushMsg)
	if err != nil {
		return 0, 0, err
	}
	// TODO Key
	m := &kafka.ProducerMessage{
		Key:   kafka.StringEncoder(strconv.FormatInt(int64(cometpb.OpRaw), 10)),
		Topic: d.topic,
		Value: kafka.ByteEncoder(b),
	}
	partition, offset, err := d.producer.SendMessage(m)
	if err != nil {
		return 0, 0, err
	}
	return partition, offset, nil
}
