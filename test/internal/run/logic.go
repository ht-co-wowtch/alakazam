package run

import (
	"github.com/DATA-DOG/go-txdb"
	"gitlab.com/jetfueltw/cpw/alakazam/activity"
	"gitlab.com/jetfueltw/cpw/alakazam/client"
	"gitlab.com/jetfueltw/cpw/alakazam/logic"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/cache"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/conf"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/grpc"
	httpServer "gitlab.com/jetfueltw/cpw/alakazam/logic/http"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/http/admin"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/http/front"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/store"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/stream"
	"net/http"
)

var Api = make(map[string]TransportFunc)

func RunLogic(path string) func() {
	if err := conf.Read(path + "/logic.yml"); err != nil {
		panic(err)
	}

	txdb.Register("mockMysql", conf.Conf.DB.Driver, store.DatabaseDns(conf.Conf.DB))
	conf.Conf.DB.Driver = "mockMysql"

	httpClient := client.Create(conf.Conf.Api, newMockClient(func(request *http.Request) (response *http.Response, e error) {
		f := Api[request.URL.Path]
		return f(request)
	}))

	c := cache.NewRedis(conf.Conf.Redis)
	srv := logic.Create(conf.Conf, store.NewStore(conf.Conf.DB), c, stream.NewKafkaPub(conf.Conf.Kafka), httpClient)

	money := activity.NewLuckyMoney(httpClient)

	httpSrv := httpServer.New(conf.Conf.HTTPServer, front.New(srv, money))
	httpAdminSrv := httpServer.New(conf.Conf.HTTPAdminServer, admin.New(srv))
	rpcSrv := grpc.New(conf.Conf.RPCServer, srv)
	return func() {
		c.FlushAll()
		srv.Close()
		httpSrv.Close()
		httpAdminSrv.Close()
		rpcSrv.GracefulStop()
	}
}

func AddClient(path string, fun TransportFunc) {
	Api[path] = fun
}

type TransportFunc func(*http.Request) (*http.Response, error)

func (tf TransportFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return tf(req)
}

func newMockClient(doer func(*http.Request) (*http.Response, error)) *http.Client {
	return &http.Client{
		Transport: TransportFunc(doer),
	}
}
