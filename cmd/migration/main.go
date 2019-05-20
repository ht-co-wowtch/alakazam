package main

import (
	"flag"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/conf"
	"gitlab.com/jetfueltw/cpw/alakazam/logic/store"
)

var (
	run      bool
	confPath string
)

func main() {
	flag.BoolVar(&run, "run", false, "run migration")
	flag.StringVar(&confPath, "c", "logic.yml", "default config path")
	flag.Parse()

	if err := conf.Read(confPath); err != nil {
		panic(err)
	}

	if run {
		store.RunMigration(".")
	}
}
