package agent

func (a *Agent) Shutdown() error {
	if err := a.members.Leave(nodeUpdateTimeout); err != nil {
		return err
	}

	if err := a.members.Shutdown(); err != nil {
		return err
	}

	return nil
}
