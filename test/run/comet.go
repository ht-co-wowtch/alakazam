package run

import (
	"gitlab.com/jetfueltw/cpw/alakazam/server/comet"
	"gitlab.com/jetfueltw/cpw/alakazam/server/comet/conf"
	"gitlab.com/jetfueltw/cpw/alakazam/server/comet/grpc"
	"math/rand"
	"runtime"
	"time"
)

func RunComet(path string) func() {
	if err := conf.Read(path + "/comet.yml"); err != nil {
		panic(err)
	}

	rand.Seed(time.Now().UTC().UnixNano())
	runtime.GOMAXPROCS(runtime.NumCPU())

	// server tcp 連線
	srv := comet.NewServer(conf.Conf)
	if err := comet.InitWebsocket(srv, conf.Conf.Websocket.Host, runtime.NumCPU()); err != nil {
		panic(err)
	}

	rpcSrv := grpc.New(conf.Conf.RPCServer, srv)

	return func() {
		rpcSrv.GracefulStop()
		srv.Close()
	}
}
