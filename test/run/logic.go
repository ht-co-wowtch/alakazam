package run

import (
	"github.com/DATA-DOG/go-txdb"
	"gitlab.com/jetfueltw/cpw/alakazam/server/logic"
	"gitlab.com/jetfueltw/cpw/alakazam/server/logic/conf"
	"gitlab.com/jetfueltw/cpw/alakazam/server/logic/grpc"
	"gitlab.com/jetfueltw/cpw/alakazam/server/logic/http"
	"gitlab.com/jetfueltw/cpw/alakazam/server/logic/http/admin"
	"gitlab.com/jetfueltw/cpw/alakazam/server/logic/http/front"
	"gitlab.com/jetfueltw/cpw/alakazam/server/logic/store"
)

func RunLogic(path string) func() {
	if err := conf.Read(path + "/logic.yml"); err != nil {
		panic(err)
	}
	txdb.Register("mockMysql", conf.Conf.DB.Driver, store.DatabaseDns(conf.Conf.DB))
	conf.Conf.DB.Driver = "mockMysql"

	srv := logic.New(conf.Conf)
	httpSrv := http.New(conf.Conf.HTTPServer, front.New(srv))
	httpAdminSrv := http.New(conf.Conf.HTTPAdminServer, admin.New(srv))
	rpcSrv := grpc.New(conf.Conf.RPCServer, srv)
	return func() {
		srv.Close()
		httpSrv.Close()
		httpAdminSrv.Close()
		rpcSrv.GracefulStop()
	}
}
