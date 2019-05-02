package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"gitlab.com/jetfueltw/cpw/alakazam/internal/job"
	"gitlab.com/jetfueltw/cpw/alakazam/internal/job/conf"
	log "github.com/golang/glog"
)

func main() {
	flag.Parse()
	if err := conf.Init(); err != nil {
		panic(err)
	}

	j := job.New(conf.Conf)
	go j.Consume()

	// 接收到close signal的處理
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		log.Infof("goim-job get a signal %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			j.Close()
			log.Flush()
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}
