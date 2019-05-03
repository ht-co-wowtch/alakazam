package main

import (
	"flag"
	"math/rand"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	log "github.com/golang/glog"
	"gitlab.com/jetfueltw/cpw/alakazam/internal/comet"
	"gitlab.com/jetfueltw/cpw/alakazam/internal/comet/conf"
	"gitlab.com/jetfueltw/cpw/alakazam/internal/comet/grpc"
)

func main() {
	flag.Parse()
	if err := conf.Init(); err != nil {
		panic(err)
	}
	rand.Seed(time.Now().UTC().UnixNano())
	runtime.GOMAXPROCS(runtime.NumCPU())

	// server tcp 連線
	srv := comet.NewServer(conf.Conf)
	if err := comet.InitWebsocket(srv, conf.Conf.Websocket.Bind, runtime.NumCPU()); err != nil {
		panic(err)
	}

	// 啟動grpc server
	rpcSrv := grpc.New(conf.Conf.RPCServer, srv)

	// 接收到close signal的處理
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		log.Infof("close get a signal %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			rpcSrv.GracefulStop()
			srv.Close()
			log.Flush()
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}
