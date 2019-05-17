package run

func Run(configPath string) func() {
	l := RunLogic(configPath)
	c := RunComet(configPath)
	j := RunJob(configPath)
	return func() {
		l()
		c()
		j()
	}
}
