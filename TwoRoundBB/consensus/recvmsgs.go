package consensus

import (
	"strconv"
	"time"
)

func (rp *Replica) GetMsg(msg interface{}) {
	switch msg.(type) {
	case *string: // for leader to begin the consensus
		rp.StartTime = time.Now().UnixNano()
		rp.ViewStartTime = rp.StartTime
		go rp.trigger()
		if rp.node.NodeID == rp.View.Leader {
			rp.propose(&Proof{})
		}
	case *ProposeMsg:
		rp.vote(msg.(*ProposeMsg))
	case *VoteMsg:
		rp.commit(msg.(*VoteMsg))
	case *CommitQC:
		rp.decide(msg.(*CommitQC))
	case *int:
		rp.removeNode(strconv.Itoa(*msg.(*int)))
	case *TimeoutMsg:
		rp.newView(msg.(*TimeoutMsg))
	case *TimeoutQC:
		rp.recvTimeoutQC(msg.(*TimeoutQC))
	case *Status:
		rp.status(msg.(*Status))
	}
}
