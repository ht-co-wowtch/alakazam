package main

import (
	"flag"
	"gitlab.com/jetfueltw/cpw/alakazam/app/comet"
	"gitlab.com/jetfueltw/cpw/alakazam/app/comet/api"
	"gitlab.com/jetfueltw/cpw/alakazam/app/comet/conf"
	"gitlab.com/jetfueltw/cpw/alakazam/cmd"
	"gitlab.com/jetfueltw/cpw/micro/log"
	"math/rand"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

var (
	// config path
	confPath string
)

func main() {
	cmd.LoadTimeZone()

	flag.StringVar(&confPath, "c", "comet.yml", "default config path.")
	flag.Parse()
	if err := conf.Read(confPath); err != nil {
		panic(err)
	}
	log.Infof("Using config file: [%s]", confPath)

	rand.Seed(time.Now().UTC().UnixNano())
	runtime.GOMAXPROCS(runtime.NumCPU())

	// server tcp 連線
	srv := comet.NewServer(conf.Conf)
	log.Infof("websocket prot [%s]", conf.Conf.Websocket.Addr)
	if err := comet.InitWebsocket(srv, conf.Conf.Websocket.Addr, runtime.NumCPU()); err != nil {
		panic(err)
	}

	// 啟動grpc server
	rpcSrv := api.New(conf.Conf.RPCServer, srv)
	log.Infof("rpc server port [%s]", conf.Conf.RPCServer.Addr)

	// 接收到close signal的處理
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		log.Infof("comet close get a signal %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			rpcSrv.GracefulStop()
			srv.Close()
			log.Sync()
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}
