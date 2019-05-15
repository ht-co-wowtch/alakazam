package run

import (
	"gitlab.com/jetfueltw/cpw/alakazam/server/job"
	"gitlab.com/jetfueltw/cpw/alakazam/server/job/conf"
)

func RunJob(path string) func() {
	if err := conf.Read(path + "/job.yml"); err != nil {
		panic(err)
	}
	j := job.New(conf.Conf)
	go j.Consume()

	return func() {
		j.Close()
	}
}
