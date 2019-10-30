package message

import (
	"encoding/json"
	"gitlab.com/jetfueltw/cpw/alakazam/app/logic/pb"
	"gitlab.com/jetfueltw/cpw/alakazam/client"
	"gitlab.com/jetfueltw/cpw/alakazam/member"
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
				if v.Type == pb.PushMsg_MONEY {
					var red RedEnvelopeMessage
					if err = json.Unmarshal(v.Msg, &red); err != nil {
						log.Error("red envelope for delay send message", zap.Error(err), zap.Int64("id", v.Seq))
						continue
					}
					if err = p.cli.UpRedEnvelopePublish(red.RedEnvelope.Id); err != nil {
						log.Error("update red envelope is publish for delay send message", zap.Error(err), zap.Int64("id", v.Seq))
						continue
					}
				}
				if err = p.producer.send(v); err != nil {
					log.Error("delay send message", zap.Error(err), zap.Int64("id", v.Seq))
				} else {
					log.Info("delay send message", zap.Int64("id", v.Seq))
				}
			}
		}
	}
}

func (p *DelayProducer) SendDelayRedEnvelopeForAdmin(msg ProducerAdminRedEnvelopeMessage, publishAt time.Time) (int64, error) {
	pushMsg, err := p.producer.toRedEnvelopePb(ProducerRedEnvelopeMessage{
		ProducerMessage: ProducerMessage{
			Rooms:   msg.Rooms,
			Mid:     member.RootMid,
			Uid:     member.RootUid,
			Name:    msg.Name,
			Message: msg.Message,
			Avatar:  99,
		},
		RedEnvelopeId: msg.RedEnvelopeId,
		Token:         msg.Token,
		Expired:       msg.Expired,
	})
	if err != nil {
		return 0, err
	}
	p.cron.add(pushMsg, publishAt)
	log.Info("add delay message for red envelope", zap.Int64("id", pushMsg.Seq))
	return pushMsg.Seq, nil
}
