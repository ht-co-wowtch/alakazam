package run

func Run(configPath string) func() {
	l := RunLogic(configPath)
	c := RunComet(configPath)
	j := RunJob(configPath)
	a := RunAdmin(configPath)
	return func() {
		l()
		c()
		j()
		a()
	}
}
