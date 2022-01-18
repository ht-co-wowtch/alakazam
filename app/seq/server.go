package seq

import (
	"context"
	"gitlab.com/ht-co/wowtch/live/alakazam/app/seq/api"
	"gitlab.com/ht-co/wowtch/live/alakazam/app/seq/conf"
	"gitlab.com/ht-co/micro/log"
	"google.golang.org/grpc"
	"net"
)

type Server struct {
	ctx       context.Context
	cancel    context.CancelFunc
	rpcServer *grpc.Server
}

func New(c *conf.Config) *Server {
	ctx, cancel := context.WithCancel(context.Background())
	srv, err := api.NewServer(ctx, c)
	lis, err := net.Listen(c.RPCServer.Network, c.RPCServer.Addr)
	if err != nil {
		panic(err)
	}

	go func() {
		if err := srv.Serve(lis); err != nil {
			log.Error(err.Error())
		}
	}()
	log.Infof("rpc server port [%s]", c.RPCServer.Addr)
	return &Server{
		ctx:       ctx,
		cancel:    cancel,
		rpcServer: srv,
	}
}

func (s *Server) Close() {
	s.cancel()
	s.rpcServer.GracefulStop()
}
