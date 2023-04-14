package node

import (
	"TwoRoundBB/config"
	"crypto/rand"
	"encoding/base64"
)

type Node struct {
	NodeID    string
	NodeTable map[string]string // key=nodeID, value=url
	NodeKey   config.NodeKey
	InitValue string // initial value of this node
}

func NewNode(nodeID string) *Node {
	b := make([]byte, config.Size) // b is a tx in bytes
	_, _ = rand.Read(b)

	node := &Node{
		NodeID:    nodeID,
		NodeTable: config.LoadNodeTable(),
		NodeKey:   config.LoadNodeKey(nodeID),
		InitValue: base64.StdEncoding.EncodeToString(b),
	}

	return node
}
