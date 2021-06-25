package main

import (
	"flag"
	"fmt"
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

	switch sig := <-c; sig {
	case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
		srv.Close()
		log.Sync()
		fmt.Println("shutdown normally")
	case syscall.SIGHUP:
	default:
		fmt.Println(sig.String())
	}
	fmt.Println("shutdown completed")
}
