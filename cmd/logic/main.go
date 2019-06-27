package main

import (
	"flag"
	"fmt"
	"gitlab.com/jetfueltw/cpw/alakazam/client"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/http/admin"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/http/front"
	"gitlab.com/jetfueltw/cpw/alakazam/models"
	"os"
	"os/signal"
	"syscall"

	log "github.com/golang/glog"
	"gitlab.com/jetfueltw/cpw/alakazam/logic"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/conf"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/grpc"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/http"
)

var (
	confPath string
	migrate  bool
)

func main() {
	flag.StringVar(&confPath, "c", "logic.yml", "default config path")
	flag.BoolVar(&migrate, "migrate", false, "run migrate")
	flag.Parse()
	if err := conf.Read(confPath); err != nil {
		panic(err)
	}
	fmt.Println("Using config file:", confPath)

	if migrate {
		if err := models.Migrate(conf.Conf.DB); err != nil {
			panic(err)
		}
		return
	}

	// new srever
	srv := logic.New(conf.Conf)
	httpSrv := http.New(conf.Conf.HTTPServer, front.New(srv, client.New(conf.Conf.Api)))
	httpAdminSrv := http.New(conf.Conf.HTTPAdminServer, admin.New(srv))
	rpcSrv := grpc.New(conf.Conf.RPCServer, srv)

	fmt.Printf("logic start success | RpcServer: %s\n", conf.Conf.RPCServer.Addr)

	// 接收到close signal的處理
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		log.Infof("logic get a signal %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			srv.Close()
			httpSrv.Close()
			httpAdminSrv.Close()
			rpcSrv.GracefulStop()
			log.Flush()
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}
