package main

import (
	"context"
	"flag"
	"fmt"
	"gitlab.com/jetfueltw/cpw/micro/log"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	addr       string
	burrowAddr string
	interval   string

	PrefixEnv = "BURROW_PROMETHEUS"
)

func main() {
	flag.StringVar(&addr, "addr", ":3037", "prometheus Address for burrow")
	flag.StringVar(&burrowAddr, "burrowAddr", "http://127.0.0.1:3400", "Address that burrow is listening on")
	flag.StringVar(&interval, "interval", "30s", "The interval(seconds) specifies how often to scrape burrow")
	flag.Parse()

	if env := os.Getenv(PrefixEnv + "_ADDR"); env != "" {
		addr = env
	}
	if env := os.Getenv(PrefixEnv + "_BURROW_ADDR"); env != "" {
		burrowAddr = env
	}
	if env := os.Getenv(PrefixEnv + "_INTERVAL"); env != "" {
		interval = env
	}

	err := log.Default()
	if err != nil {
		fmt.Println("log error %v", err)
	}

	scrape, err := time.ParseDuration(interval)
	if err != nil {
		log.Error("interval(seconds) ", zap.Error(err))
		return
	}

	ctx, cancel := context.WithCancel(context.Background())

	e, err := newExporter(ctx, addr, burrowAddr, scrape)
	if err != nil {
		log.Error("start failure", zap.Error(err))
		return
	}

	e.Start()

	log.Info("start burrow exporter")

	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)

	<-done
	cancel()
	log.Info("shutting down exporter")
}
