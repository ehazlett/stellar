package agent

import "time"

func (a *Agent) Shutdown() error {
	if err := a.members.Leave(time.Second * 1); err != nil {
		return err
	}

	return nil
}
