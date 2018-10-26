package element

// Shutdown causes the local node to leave the cluster and perform a clean shutdown
func (a *Agent) Shutdown() error {
	if err := a.members.Leave(nodeUpdateTimeout); err != nil {
		return err
	}
	return a.members.Shutdown()
}
