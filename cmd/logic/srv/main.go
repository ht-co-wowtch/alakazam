package main

/*
這個 main 只開啟 HTTPServer
*/
import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	//"github.com/Terry-Mao/goim/internal/logic"
	"gitlab.com/jetfueltw/cpw/alakazam/internal/logic/conf"
	//"github.com/Terry-Mao/goim/internal/logic/grpc"
	log "github.com/golang/glog"
	"gitlab.com/jetfueltw/cpw/alakazam/internal/logic/http"
)

func main() {
	flag.Parse()
	if err := conf.Init(); err != nil {
		panic(err)
	}

	// new srever
	//srv := logic.New(conf.Conf)
	//httpSrv := http.New(conf.Conf.HTTPServer, srv)
	httpSrv := http.New(conf.Conf.HTTPServer, nil)
	//rpcSrv := grpc.New(conf.Conf.RPCServer, srv)

	// 接收到close signal的處理
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		log.Infof("goim-logic get a signal %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			//srv.Close()
			httpSrv.Close()
			//rpcSrv.GracefulStop()
			log.Flush()
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}
