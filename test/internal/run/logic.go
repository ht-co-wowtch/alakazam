package run

import (
	"bytes"
	"encoding/json"
	"github.com/DATA-DOG/go-txdb"
	"gitlab.com/jetfueltw/cpw/alakazam/logic"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/cache"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/client"
	user "gitlab.com/jetfueltw/cpw/alakazam/logic/client"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/conf"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/grpc"
	httpServer "gitlab.com/jetfueltw/cpw/alakazam/logic/http"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/http/admin"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/http/front"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/store"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/stream"
	"io/ioutil"
	"net/http"
)

func RunLogic(path string) func() {
	if err := conf.Read(path + "/logic.yml"); err != nil {
		panic(err)
	}
	txdb.Register("mockMysql", conf.Conf.DB.Driver, store.DatabaseDns(conf.Conf.DB))
	conf.Conf.DB.Driver = "mockMysql"

	httpClient := client.Create(conf.Conf.Api, newMockClient(func(request *http.Request) (response *http.Response, e error) {
		u := user.User{
			Uid:  "82ea16cd2d6a49d887440066ef739669",
			Name: "test",
		}

		b, err := json.Marshal(u)
		if err != nil {
			return nil, err
		}

		header := http.Header{}
		header.Set("Content-Type", "application/json")

		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader(b)),
			Header:     header,
		}, nil
	}))

	c := cache.NewRedis(conf.Conf.Redis)
	srv := logic.Create(conf.Conf, store.NewStore(conf.Conf.DB), c, stream.NewKafkaPub(conf.Conf.Kafka), httpClient)
	httpSrv := httpServer.New(conf.Conf.HTTPServer, front.New(srv))
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

type transportFunc func(*http.Request) (*http.Response, error)

func (tf transportFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return tf(req)
}

func newMockClient(doer func(*http.Request) (*http.Response, error)) *http.Client {
	return &http.Client{
		Transport: transportFunc(doer),
	}
}
