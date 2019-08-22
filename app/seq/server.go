package seq

import (
	"context"
	"gitlab.com/jetfueltw/cpw/alakazam/app/seq/api"
	"gitlab.com/jetfueltw/cpw/alakazam/app/seq/conf"
	"gitlab.com/jetfueltw/cpw/micro/log"
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
			panic(err)
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
