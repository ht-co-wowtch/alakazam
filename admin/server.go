package admin

import (
	"context"
	goRedis "github.com/go-redis/redis"
	"gitlab.com/jetfueltw/cpw/alakazam/admin/api"
	"gitlab.com/jetfueltw/cpw/alakazam/admin/conf"
	"gitlab.com/jetfueltw/cpw/alakazam/client"
	"gitlab.com/jetfueltw/cpw/alakazam/member"
	"gitlab.com/jetfueltw/cpw/alakazam/message"
	"gitlab.com/jetfueltw/cpw/alakazam/models"
	"gitlab.com/jetfueltw/cpw/alakazam/room"
	"gitlab.com/jetfueltw/cpw/micro/log"
	"gitlab.com/jetfueltw/cpw/micro/redis"
	"net/http"
)

type Server struct {
	ctx        context.Context
	cancel     context.CancelFunc
	cache      *goRedis.Client
	httpServer *http.Server
}

func New(conf *conf.Config) *Server {
	ctx, cancel := context.WithCancel(context.Background())
	cache := redis.New(conf.Redis)
	db := models.NewStore(conf.DB)
	cli := client.New(conf.Nidoran)
	messageProducer := message.NewProducer(conf.Kafka.Brokers, conf.Kafka.Topic)
	memberCli := member.New(db, cache, cli)
	roomCli := room.New(db, cache, memberCli, cli, 0)
	httpServer := api.NewServer(conf.HTTPServer, memberCli, messageProducer, roomCli, cli)
	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()
	log.Infof("http server port [%s]", conf.HTTPServer.Addr)

	return &Server{
		ctx:        ctx,
		cancel:     cancel,
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
