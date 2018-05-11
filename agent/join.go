package agent

import "github.com/sirupsen/logrus"

func (a *Agent) Join(peers []string) error {
	n, err := a.members.Join(peers)
	if err != nil {
		return err
	}

	logrus.Debugf("joined %d peer(s)", n)
	return nil
}
