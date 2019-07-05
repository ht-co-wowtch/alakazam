package main

import (
	"flag"
	"fmt"
	"gitlab.com/jetfueltw/cpw/alakazam/admin/conf"
	admin "gitlab.com/jetfueltw/cpw/alakazam/admin/http"
	"gitlab.com/jetfueltw/cpw/alakazam/logic"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/http"
	"gitlab.com/jetfueltw/cpw/micro/log"
	"os"
	"os/signal"
	"syscall"
)

var (
	confPath string
)

func main() {
	flag.StringVar(&confPath, "c", "admin.yml", "default config path")
	flag.Parse()
	if err := conf.Read(confPath); err != nil {
		panic(err)
	}
	fmt.Println("Using config file:", confPath)

	srv := logic.NewAdmin(conf.Conf.DB, conf.Conf.Redis, conf.Conf.Kafka)
	httpAdminSrv := http.New(conf.Conf.HTTPServer, admin.New(srv))

	fmt.Printf("admin start success")

	// 接收到close signal的處理
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		log.Infof("logic get a signal %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			srv.Close()
			httpAdminSrv.Close()
			log.Sync()
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}
