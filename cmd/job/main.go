package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	log "github.com/golang/glog"
	"gitlab.com/jetfueltw/cpw/alakazam/server/job"
	"gitlab.com/jetfueltw/cpw/alakazam/server/job/conf"
)

func main() {
	flag.Parse()
	if err := conf.Init(); err != nil {
		panic(err)
	}

	j := job.New(conf.Conf)
	go j.Consume()

	fmt.Println("job start success")

	// 接收到close signal的處理
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		log.Infof("job close get a signal %s", s.String())
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
