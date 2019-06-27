package run

import (
	"gitlab.com/jetfueltw/cpw/alakazam/activity"
	"gitlab.com/jetfueltw/cpw/alakazam/client"
	"gitlab.com/jetfueltw/cpw/alakazam/logic"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/conf"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/grpc"
	httpServer "gitlab.com/jetfueltw/cpw/alakazam/logic/http"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/http/admin"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/http/front"
)

func RunLogic(path string) func() {
	if err := conf.Read(path + "/logic.yml"); err != nil {
		panic(err)
	}

	srv := logic.New(conf.Conf)
	money := activity.NewRedEnvelope(client.New(conf.Conf.Api))
	httpSrv := httpServer.New(conf.Conf.HTTPServer, front.New(srv, money))
	httpAdminSrv := httpServer.New(conf.Conf.HTTPAdminServer, admin.New(srv))
	rpcSrv := grpc.New(conf.Conf.RPCServer, srv)

	return func() {
		srv.Close()
		httpSrv.Close()
		httpAdminSrv.Close()
		rpcSrv.GracefulStop()
	}
}
