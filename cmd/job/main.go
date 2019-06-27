package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	log "github.com/golang/glog"
	"gitlab.com/jetfueltw/cpw/alakazam/job"
	"gitlab.com/jetfueltw/cpw/alakazam/job/conf"
)

var (
	// config path
	confPath string
)

func main() {
	flag.StringVar(&confPath, "c", "job-example.yml", "default config path")
	flag.Parse()
	if err := conf.Read(confPath); err != nil {
		panic(err)
	}
	fmt.Println("Using config file:", confPath)

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
