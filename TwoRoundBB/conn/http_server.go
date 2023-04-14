package conn

import (
	"TwoRoundBB/config"
	"TwoRoundBB/consensus"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

func (srv *Srv) HTTPStart() {
	id, _ := strconv.Atoi(srv.node.NodeID)
	fmt.Printf("Total: %v replicas, with %v faulty nodes.\n", config.N, config.F)
	fmt.Printf("Network delay: %v (%v) ms , quorum size: %v, tx size: %v bytes.\n", config.T, config.T+id%config.T, config.QC2, config.Size)
	fmt.Printf("Timer begin:%v ms\n", config.D)
	fmt.Println()
	fmt.Printf("Server will be started at %s...\n", srv.url)
	if err := http.ListenAndServe(srv.url, nil); err != nil {
		fmt.Println(err)
		return
	}
}

var client *http.Client

// pick to be sent msg from channel: MsgSendChan
func (srv *Srv) send() {
	client = &http.Client{}
	client.Transport = &http.Transport{
		MaxIdleConnsPerHost: 100,
	}
	for {
		sendmsg := <-srv.MsgSendChan
		buff := bytes.NewBuffer(sendmsg.msg)
		//fmt.Println("post to", sendmsg.url)
		_, err := client.Post("http://"+sendmsg.url, "application/json", buff)
		if err != nil {
			fmt.Println("wrong when send msg:", err)
		}
		time.Sleep(time.Microsecond)
	}
}

// http receive functions
func (srv *Srv) setRoute() {
	http.HandleFunc("/begin", srv.getBegin)
	http.HandleFunc("/propose", srv.getPropose)
	http.HandleFunc("/vote", srv.getVote)
	http.HandleFunc("/commitQC", srv.getCommitQC)
	http.HandleFunc("/done", srv.getDone)

	http.HandleFunc("/timeout", srv.getTimeout)
	http.HandleFunc("/timeoutQC", srv.getTimeoutQC)
	http.HandleFunc("/status", srv.getStatus)
}

func (srv *Srv) getBegin(writer http.ResponseWriter, request *http.Request) {
	fmt.Println("[State]: begin consensus!")
	var msg string
	err := json.NewDecoder(request.Body).Decode(&msg)
	if err != nil {
		fmt.Println("json.NewDecoder err:", err)
		return
	}
	srv.MsgEntrance <- &msg
}

func (srv *Srv) getPropose(writer http.ResponseWriter, request *http.Request) {
	var msg consensus.ProposeMsg
	err := json.NewDecoder(request.Body).Decode(&msg)

	if err != nil {
		fmt.Println(err)
		return
	}
	srv.MsgEntrance <- &msg
}

func (srv *Srv) getVote(writer http.ResponseWriter, request *http.Request) {
	var msg consensus.VoteMsg
	err := json.NewDecoder(request.Body).Decode(&msg)
	if err != nil {
		fmt.Println(err)
		return
	}
	srv.MsgEntrance <- &msg
}

func (srv *Srv) getCommitQC(writer http.ResponseWriter, request *http.Request) {
	var msg consensus.CommitQC
	err := json.NewDecoder(request.Body).Decode(&msg)
	if err != nil {
		fmt.Println(err)
		return
	}
	srv.MsgEntrance <- &msg
}

func (srv *Srv) getDone(writer http.ResponseWriter, request *http.Request) {
	var msg int
	err := json.NewDecoder(request.Body).Decode(&msg)
	if err != nil {
		fmt.Println(err)
		return
	}
	srv.MsgEntrance <- &msg
}

func (srv *Srv) getTimeout(writer http.ResponseWriter, request *http.Request) {
	var msg consensus.TimeoutMsg
	err := json.NewDecoder(request.Body).Decode(&msg)
	if err != nil {
		fmt.Println(err)
		return
	}
	srv.MsgEntrance <- &msg
}

func (srv *Srv) getTimeoutQC(writer http.ResponseWriter, request *http.Request) {
	var msg consensus.TimeoutQC
	err := json.NewDecoder(request.Body).Decode(&msg)
	if err != nil {
		fmt.Println(err)
		return
	}
	srv.MsgEntrance <- &msg
}

func (srv *Srv) getStatus(writer http.ResponseWriter, request *http.Request) {
	var msg consensus.Status
	err := json.NewDecoder(request.Body).Decode(&msg)
	if err != nil {
		fmt.Println(err)
		return
	}
	srv.MsgEntrance <- &msg
}
