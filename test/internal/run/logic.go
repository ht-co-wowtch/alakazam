package run

import (
	"gitlab.com/jetfueltw/cpw/alakazam/client"
	"gitlab.com/jetfueltw/cpw/alakazam/logic"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/conf"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/grpc"
	httpServer "gitlab.com/jetfueltw/cpw/alakazam/logic/http"
)

func RunLogic(path string) func() {
	if err := conf.Read(path + "/logic.yml"); err != nil {
		panic(err)
	}

	srv := logic.New(conf.Conf)
	httpSrv := httpServer.New(
		conf.Conf.HTTPServer,
		httpServer.NewContext(
			srv,
			client.New(conf.Conf.Api),
		),
	)
	rpcSrv := grpc.New(conf.Conf.RPCServer, srv)

	return func() {
		srv.Close()
		httpSrv.Close()
		rpcSrv.GracefulStop()
	}
}
