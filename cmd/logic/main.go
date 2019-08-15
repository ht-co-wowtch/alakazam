package main

import (
	"flag"
	"gitlab.com/jetfueltw/cpw/alakazam/app/logic"
	"gitlab.com/jetfueltw/cpw/alakazam/app/logic/conf"
	"gitlab.com/jetfueltw/cpw/alakazam/models"
	"gitlab.com/jetfueltw/cpw/micro/log"
	"os"
	"os/signal"
	"syscall"
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
	log.Infof("Using config file: [%s]", confPath)

	if migrate {
		if err := models.Migrate(conf.Conf.DB); err != nil {
			panic(err)
		}
		return
	}

	srv := logic.New(conf.Conf)

	// 接收到close signal的處理
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		log.Infof("logic get a signal %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			srv.Close()
			log.Sync()
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}
