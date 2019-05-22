package run

import (
	"github.com/DATA-DOG/go-txdb"
	"gitlab.com/jetfueltw/cpw/alakazam/logic"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/cache"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/conf"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/grpc"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/http"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/http/admin"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/http/front"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/store"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/stream"
)

func RunLogic(path string) func() {
	if err := conf.Read(path + "/logic.yml"); err != nil {
		panic(err)
	}
	txdb.Register("mockMysql", conf.Conf.DB.Driver, store.DatabaseDns(conf.Conf.DB))
	conf.Conf.DB.Driver = "mockMysql"

	c := cache.NewRedis(conf.Conf.Redis)
	srv := logic.Create(conf.Conf, store.NewStore(conf.Conf.DB), c, stream.NewKafkaPub(conf.Conf.Kafka))
	httpSrv := http.New(conf.Conf.HTTPServer, front.New(srv))
	httpAdminSrv := http.New(conf.Conf.HTTPAdminServer, admin.New(srv))
	rpcSrv := grpc.New(conf.Conf.RPCServer, srv)
	return func() {
		c.FlushAll()
		srv.Close()
		httpSrv.Close()
		httpAdminSrv.Close()
		rpcSrv.GracefulStop()
	}
}
