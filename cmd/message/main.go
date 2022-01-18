package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"gitlab.com/ht-co/wowtch/live/alakazam/app/message"
	"gitlab.com/ht-co/wowtch/live/alakazam/app/message/conf"
	"gitlab.com/ht-co/wowtch/live/alakazam/cmd"
	"gitlab.com/ht-co/wowtch/live/alakazam/pkg/metrics"
	"gitlab.com/ht-co/micro/log"
)

var (
	// config path
	confPath string
)

func main() {
	cmd.LoadTimeZone()

	flag.StringVar(&confPath, "c", "message.yml", "default config path")
	flag.Parse()
	if err := conf.Read(confPath); err != nil {
		panic(err)
	}
	log.Infof("Using config file: [%s]", confPath)

	m := message.New(conf.Conf)
	m.Run()
	metrics.RunHttp(conf.Conf.MetricsAddr)

	log.Info("start success")
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)

	switch sig := <-c; sig {
	case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
		m.Close()
		log.Sync()
		fmt.Println("shutdown normally")
	case syscall.SIGHUP:
	default:
		fmt.Println(sig.String())
	}
	fmt.Println("shutdown completed")
}
