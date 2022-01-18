package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"gitlab.com/ht-co/wowtch/live/alakazam/app/comet"
	"gitlab.com/ht-co/wowtch/live/alakazam/app/comet/api"
	"gitlab.com/ht-co/wowtch/live/alakazam/app/comet/conf"
	"gitlab.com/ht-co/wowtch/live/alakazam/cmd"
	"gitlab.com/ht-co/wowtch/live/alakazam/pkg/metrics"
	"gitlab.com/ht-co/cpw/micro/log"
)

var (
	// config path
	confPath string
)

func main() {
	cmd.LoadTimeZone()

	// 取得comet相關設定
	flag.StringVar(&confPath, "c", "comet.yml", "default config path.")
	flag.Parse()
	if err := conf.Read(confPath); err != nil {
		panic(err)
	}

	rand.Seed(time.Now().UTC().UnixNano())
	runtime.GOMAXPROCS(runtime.NumCPU())

	// server tcp 連線
	srv := comet.NewServer(conf.Conf)
	//log.Infof("websocket prot [%s]", conf.Conf.Websocket.Addr)

	// 建立websocket service
	if err := comet.InitWebsocket(srv, conf.Conf.Websocket.Addr, runtime.NumCPU()); err != nil {
		panic(err)
	}

	// 啟動grpc server
	rpcSrv := api.New(conf.Conf.RPCServer, srv)
	//log.Infof("rpc server port [%s]", conf.Conf.RPCServer.Addr)

	metrics.RunHttp(conf.Conf.MetricsAddr)

	// 接收到close signal的處理
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)

	switch sig := <-c; sig {
	case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
		rpcSrv.GracefulStop()
		srv.Close()
		log.Sync()
		fmt.Println("shutdown normally")
	case syscall.SIGHUP:
	default:
		fmt.Println(sig.String())
	}
	fmt.Println("shutdown completed")
}
