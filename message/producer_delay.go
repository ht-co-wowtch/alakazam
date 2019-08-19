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
				if err := p.producer.send(v); err != nil {
					log.Error("delay send message", zap.Int64("id", v.Seq))
				} else {
					log.Info("delay send message", zap.Int64("id", v.Seq))
				}
			}
		}
	}
}

func (p *DelayProducer) SendDelayRedEnvelopeForAdmin(msg AdminRedEnvelopeMessage, publishAt time.Time) error {
	pushMsg, err := p.producer.toRedEnvelopePb(RedEnvelopeMessage{
		Messages: Messages{
			Rooms:   msg.Rooms,
			Rids:    msg.Rids,
			Mid:     RootMid,
			Uid:     RootUid,
			Name:    RootName,
			Message: msg.Message,
		},
		RedEnvelopeId: msg.RedEnvelopeId,
		Token:         msg.Token,
		Expired:       msg.Expired,
	})
	if err != nil {
		return err
	}
	p.cron.add(pushMsg, publishAt)
	log.Info("add delay message for red envelope", zap.Int64("id", pushMsg.Seq))
	return nil
}
