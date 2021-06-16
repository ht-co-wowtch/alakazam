package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"gitlab.com/jetfueltw/cpw/alakazam/app/logic"
	"gitlab.com/jetfueltw/cpw/alakazam/app/logic/conf"
	"gitlab.com/jetfueltw/cpw/alakazam/cmd"
	"gitlab.com/jetfueltw/cpw/alakazam/pkg/metrics"
	"gitlab.com/jetfueltw/cpw/micro/log"

	// "net/http/pprof"
	// "runtime/pprof"
)

var (
	confPath string
)

func main() {
	cmd.LoadTimeZone()

	flag.StringVar(&confPath, "c", "logic.yml", "default config path")
	flag.Parse()
	if err := conf.Read(confPath); err != nil {
		panic(err)
	}
	log.Infof("Using config file: [%s]", confPath)

	srv := logic.New(conf.Conf)
	metrics.RunHttp(conf.Conf.MetricsAddr)

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
