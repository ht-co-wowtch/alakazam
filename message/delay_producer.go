package message

import (
	"gitlab.com/jetfueltw/cpw/alakazam/client"
	"gitlab.com/jetfueltw/cpw/micro/log"
	"go.uber.org/zap"
	"time"
)

type DelayProducer struct {
	producer *Producer
	cli      *client.Client
	cron     *cron
	stop     chan struct{}
}

func NewDelayProducer(producer *Producer, cli *client.Client) *DelayProducer {
	return &DelayProducer{
		producer: producer,
		cli:      cli,
		cron:     newCron(time.Second * 5),
		stop:     make(chan struct{}),
	}
}

func (p *DelayProducer) Start() {
	go p.cron.run()
	go p.run()
}

func (p *DelayProducer) Close() error {
	err := p.producer.Close()
	p.cron.close()
	p.stop <- struct{}{}
	return err
}

func (p *DelayProducer) run() {
	for {
		select {
		case <-p.stop:
			log.Info("close delay producer")
			return
		case m := <-p.cron.Message():
			for _, v := range m {
				var err error
				switch v.category {
				case message_category:
					//err = p.producer.Send(v.room[0])
				case redenvelope_message_category:
					//err = p.producer.SendRedEnvelope(v.room[0], v.message, v.redEnvelope)
				}
				if err != nil {
					log.Error("delay send message", zap.Int64("id", v.message.Id))
				} else {
					log.Info("delay send message", zap.Int64("id", v.message.Id))
				}
			}
		}
	}
}

func (p *DelayProducer) SendDelayRedEnvelope(roomId string, message Message, envelope RedEnvelope, publishAt time.Time) error {
	p.cron.add(messageSet{
		room:        []string{roomId},
		message:     message,
		redEnvelope: envelope,
		category:    redenvelope_message_category,
	}, publishAt)
	log.Info("add delay message for red envelope", zap.Int64("id", message.Id))
	return nil
}
