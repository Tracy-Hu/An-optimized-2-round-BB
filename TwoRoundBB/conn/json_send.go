package conn

import (
	"TwoRoundBB/config"
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

func (srv *Srv) JsonSendToLeader(msg interface{}, url string, path string) {
	id, _ := strconv.Atoi(srv.node.NodeID)
	time.Sleep(time.Duration(config.T+id%config.T) * time.Millisecond) // simulate network delay

	jsonMsg, err := json.Marshal(msg)
	if err != nil {
		fmt.Println(err)
	}
	m := &sendMsg{
		url: url + path,
		msg: jsonMsg,
	}
	srv.MsgSendChan <- *m
}

// JsonSendToAll does not send to the node itself
func (srv *Srv) JsonSendToAll(msg interface{}, path string) {
	id, _ := strconv.Atoi(srv.node.NodeID)
	time.Sleep(time.Duration(config.T+id%config.T) * time.Millisecond) // simulate network delay

	jsonMsg, err := json.Marshal(msg)
	if err != nil {
		fmt.Println(err)
	}
	for nodeID, url := range srv.node.NodeTable {
		//fmt.Printf("send to %v: ", nodeID)
		if nodeID == srv.node.NodeID { // ignore itself
			//fmt.Println("ok,next.")
			continue
		}

		m := &sendMsg{
			url: url + path,
			msg: jsonMsg,
		}
		srv.MsgSendChan <- *m
	}
	return
}

// JsonSendToSome send to some specific nodes
func (srv *Srv) JsonSendToSome(msg interface{}, path string, nodeid []string) {
	id, _ := strconv.Atoi(srv.node.NodeID)
	time.Sleep(time.Duration(config.T+id%config.T) * time.Millisecond) // simulate network delay

	jsonMsg, err := json.Marshal(msg)
	if err != nil {
		fmt.Println(err)
	}
	for _, nodeID := range nodeid {
		fmt.Printf("send to %v,", nodeID)

		m := &sendMsg{
			url: srv.node.NodeTable[nodeID] + path,
			msg: jsonMsg,
		}
		srv.MsgSendChan <- *m
	}
	fmt.Println()
	return
}
