package admin

import (
	"context"
	"net/http"

	goRedis "github.com/go-redis/redis"
	"gitlab.com/ht-co/wowtch/live/alakazam/app/admin/api"
	"gitlab.com/ht-co/wowtch/live/alakazam/app/admin/conf"
	seqpb "gitlab.com/ht-co/wowtch/live/alakazam/app/seq/api/pb"
	"gitlab.com/ht-co/wowtch/live/alakazam/client"
	"gitlab.com/ht-co/wowtch/live/alakazam/member"
	"gitlab.com/ht-co/wowtch/live/alakazam/message"
	"gitlab.com/ht-co/wowtch/live/alakazam/models"
	"gitlab.com/ht-co/wowtch/live/alakazam/room"
	rpccli "gitlab.com/ht-co/micro/grpc"
	"gitlab.com/ht-co/micro/log"
	"gitlab.com/ht-co/micro/redis"
	//_ "net/http/pprof"
)

type Server struct {
	message      *message.Producer
	delayMessage *message.DelayProducer
	cancel       context.CancelFunc
	cache        *goRedis.Client
	httpServer   *http.Server
	ctx          context.Context
}

func New(conf *conf.Config) *Server {
	ctx, cancel := context.WithCancel(context.Background())
	cache := redis.New(conf.Redis)
	db := models.NewStore(conf.DB)
	cli := client.New(conf.Nidoran)
	seqCli, err := rpccli.NewClient(conf.Seq)
	if err != nil {
		panic(err)
	}
	messageProducer := message.NewProducer(conf.Kafka.Brokers, conf.Kafka.Topic, seqpb.NewSeqClient(seqCli), cache, db)
	shield := message.NewFilter(db)
	memberCli := member.New(db, cache, cli)
	roomCli := room.New(db, cache, memberCli, cli)
	httpServer := api.NewServer(conf.HTTPServer, memberCli, messageProducer, roomCli, cli, shield)
	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()
	log.Infof("http server port [%s]", conf.HTTPServer.Addr)

	return &Server{
		ctx:        ctx,
		cancel:     cancel,
		cache:      cache,
		httpServer: httpServer,
	}
}

func (s *Server) Close() {
	if err := s.cache.Close(); err != nil {
		log.Errorf("redis close error(%v)", err)
	}
	if err := s.httpServer.Shutdown(s.ctx); err != nil {
		log.Errorf("http server close error(%v)", err)
	}
}
