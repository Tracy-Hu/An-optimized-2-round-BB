package consensus

import (
	"TwoRoundBB/crypto"
	"TwoRoundBB/node"
	"TwoRoundBB/types"
	"fmt"
)

type View struct {
	ViewNum int
	Leader  string
}

type Replica struct {
	node             *node.Node
	View             *View
	ValidVoteMsgs    []*VoteMsg
	ValidTimeoutMsgs map[int][]*TimeoutMsg
	ValidStatusMsgs  []*Status
	HighQC           *TimeoutQC
	ConnI            types.ConnInterface
	StartTime        int64
	ViewStartTime    int64
	State            int // 1: normal-case; 2: view-chang; 3: commit.

	// for view-change
	Voted *VoteMsg
}

func NewReplica(node *node.Node) *Replica {
	// initialization
	view := &View{
		ViewNum: 1,
		Leader:  "1",
	}
	highQC := &TimeoutQC{
		TQC:       nil,
		ViewNum:   -1,
		LockType:  0,
		LockValue: "",
	}
	voted := &VoteMsg{
		NodeID:     "",
		Value:      "",
		ViewNum:    -1,
		LeaderID:   "",
		SignLeader: crypto.ECDSAsign{},
		SignIn:     crypto.ECDSAsign{},
		SignOut:    crypto.ECDSAsign{},
	}

	rp := &Replica{
		node:             node,
		View:             view,
		ValidVoteMsgs:    make([]*VoteMsg, 0),
		ValidTimeoutMsgs: make(map[int][]*TimeoutMsg, 0),
		ValidStatusMsgs:  make([]*Status, 0),
		HighQC:           highQC,
		State:            1,
		Voted:            voted,
	}

	fmt.Printf("[State] view number: %v, view leader: node %v\n", view.ViewNum, view.Leader)
	fmt.Printf("I'm node %v!\n", node.NodeID)

	return rp
}
