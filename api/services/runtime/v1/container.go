package runtime

func (c *Container) Running() bool {
	return c.Task.Pid > 0
}
