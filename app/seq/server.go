package seq

import (
	"gitlab.com/jetfueltw/cpw/alakazam/app/seq/api"
	"gitlab.com/jetfueltw/cpw/alakazam/app/seq/conf"
	"gitlab.com/jetfueltw/cpw/micro/log"
	"google.golang.org/grpc"
	"net"
)

type Server struct {
	rpcServer *grpc.Server
}

func New(c *conf.Config) *Server {
	srv, err := api.NewServer(c)
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
	return &Server{rpcServer: srv}
}

func (s *Server) Close() {
	s.rpcServer.GracefulStop()
}
