package main

import (
	"flag"
	"gitlab.com/jetfueltw/cpw/alakazam/app/message"
	"gitlab.com/jetfueltw/cpw/alakazam/app/message/conf"
	"gitlab.com/jetfueltw/cpw/alakazam/cmd"
	"gitlab.com/jetfueltw/cpw/alakazam/pkg/metrics"
	"gitlab.com/jetfueltw/cpw/micro/log"
	"os"
	"os/signal"
	"syscall"
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

	// 接收到close signal的處理
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		log.Infof("message close get a signal %s", s.String())
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			m.Close()
			log.Sync()
			return
		default:
			return
		}
	}
}
