package conn

import (
	"TwoRoundBB/node"
	"TwoRoundBB/types"
)

type sendMsg struct {
	url string
	msg []byte
}

type Srv struct {
	url         string
	node        *node.Node
	MsgSendChan chan sendMsg
	MsgEntrance chan interface{}

	ConsensusI types.ConsensusInterface
}

func NewServer(nodeID string) (*Srv, *node.Node) {
	node := node.NewNode(nodeID)
	server := &Srv{
		url:         node.NodeTable[nodeID],
		node:        node,
		MsgSendChan: make(chan sendMsg, 200),
		MsgEntrance: make(chan interface{}),
	}
	server.setRoute()

	go server.DispatchMsg()
	go server.send()

	return server, node
}

func (srv *Srv) DispatchMsg() {
	for {
		msg := <-srv.MsgEntrance
		//fmt.Println("MsgEntrance get msg")
		srv.ConsensusI.GetMsg(msg)
		//time.Sleep(time.Microsecond)
	}
}
