package logic

import (
	"context"
	goRedis "github.com/go-redis/redis"
	api "gitlab.com/jetfueltw/cpw/alakazam/app/logic/api/http"
	"gitlab.com/jetfueltw/cpw/alakazam/app/logic/api/rpc"
	"gitlab.com/jetfueltw/cpw/alakazam/app/logic/conf"
	seqpb "gitlab.com/jetfueltw/cpw/alakazam/app/seq/api/pb"
	"gitlab.com/jetfueltw/cpw/alakazam/client"
	"gitlab.com/jetfueltw/cpw/alakazam/member"
	"gitlab.com/jetfueltw/cpw/alakazam/message"
	"gitlab.com/jetfueltw/cpw/alakazam/models"
	"gitlab.com/jetfueltw/cpw/alakazam/room"
	rpccli "gitlab.com/jetfueltw/cpw/micro/grpc"
	"gitlab.com/jetfueltw/cpw/micro/log"
	"gitlab.com/jetfueltw/cpw/micro/redis"
	"google.golang.org/grpc"
	"net"
	"net/http"
	"time"
)

const (
	onlineTick = time.Second * 30
)

type Server struct {
	ctx        context.Context
	cancel     context.CancelFunc
	db         *models.Store
	cache      *goRedis.Client
	message    *message.Producer
	client     *client.Client
	member     *member.Member
	room       *room.Room
	httpServer *http.Server
	rpc        *grpc.Server

	// 房間在線人數，key是房間id
	roomCount map[string]int32
}

func New(c *conf.Config) *Server {
	ctx, cancel := context.WithCancel(context.Background())
	cache := redis.New(c.Redis)
	db := models.NewStore(c.DB)
	cli := client.New(c.Nidoran)
	seqCli, err := rpccli.NewClient(c.Seq)
	if err != nil {
		panic(err)
	}
	messageProducer := message.NewProducer(c.Kafka.Brokers, c.Kafka.Topic, seqpb.NewSeqClient(seqCli), cache, db)
	memberCli := member.New(db, cache, cli)
	roomCli := room.New(db, cache, memberCli, cli, c.Heartbeat)
	httpServer := api.NewServer(c.HTTPServer, memberCli, messageProducer, roomCli, cli, message.NewHistory(db, memberCli))
	rpcServer := rpc.New(c.RPCServer, roomCli)
	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(err)
		}
	}()
	log.Infof("http server port [%s]", c.HTTPServer.Addr)

	lis, err := net.Listen(c.RPCServer.Network, c.RPCServer.Addr)
	if err != nil {
		panic(err)
	}
	go func() {
		if err := rpcServer.Serve(lis); err != nil {
			panic(err)
		}
	}()
	log.Infof("rpc server port [%s]", c.RPCServer.Addr)

	s := &Server{
		ctx:        ctx,
		cancel:     cancel,
		db:         db,
		cache:      cache,
		message:    messageProducer,
		client:     cli,
		member:     memberCli,
		room:       roomCli,
		httpServer: httpServer,
		rpc:        rpcServer,
	}

	_ = s.loadOnline()
	go s.onlineproc()
	return s
}

func (s *Server) Close() {
	if err := s.cache.Close(); err != nil {
		log.Errorf("redis close error(%v)", err)
	}
	if err := s.message.Close(); err != nil {
		log.Errorf("message producer close error(%v)", err)
	}
	if err := s.httpServer.Shutdown(s.ctx); err != nil {
		log.Errorf("http server close error(%v)", err)
	}
	s.rpc.GracefulStop()
}

func (s *Server) onlineproc() {
	for {
		time.Sleep(onlineTick)
		if err := s.loadOnline(); err != nil {
			log.Errorf("onlineproc error(%v)", err)
		}
	}
}

// 從redis拿出現在各房間人數
func (s *Server) loadOnline() (err error) {
	var (
		roomCount = make(map[string]int32)
	)
	var online *room.Online
	// TODO hostname 先寫死 後續需要註冊中心來sync
	online, err = s.room.GetOnline("hostname")
	if err != nil {
		return
	}

	for roomID, count := range online.RoomCount {
		roomCount[roomID] += count
	}
	s.roomCount = roomCount
	return
}
