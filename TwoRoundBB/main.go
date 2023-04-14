package main

import (
	"TwoRoundBB/conn"
	"TwoRoundBB/consensus"
	"os"
)

// server is a name for communication;
// node is a name for configuration;
// replica is a name for consensus.
func main() {
	nodeID := os.Args[1]
	server, node := conn.NewServer(nodeID)
	rp := consensus.NewReplica(node)
	server.ConsensusI = rp
	rp.ConnI = server

	server.HTTPStart()
}
