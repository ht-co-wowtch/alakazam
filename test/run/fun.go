package run

func Run(path string) func() {
	l := RunLogic(path)
	c := RunComet(path)
	j := RunJob(path)
	return func() {
		l()
		c()
		j()
	}
}
