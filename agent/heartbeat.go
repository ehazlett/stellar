package agent

import (
	"fmt"
)

func (a *Agent) heartbeat() {
	self := a.members.LocalNode()
	fmt.Printf("meta: %s\n", string(self.Meta))
	//for _, peer := range a.peers {
	//	ac, err := NewAgentClient(remoteAgent.Addr)
	//	if err != nil {
	//		logrus.Errorf("error communicating with peer: %s", err)
	//		return
	//	}

	//	v, err := ac.VersionService.Version(context.Background(), nil)
	//	if err != nil {
	//		logrus.Error(err)
	//	}
	//}
}
