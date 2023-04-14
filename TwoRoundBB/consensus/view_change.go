package consensus

import (
	"TwoRoundBB/config"
	"TwoRoundBB/crypto"
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

// leader election def
func getViewLeader(w int) string {
	return strconv.Itoa((w-1)%config.N + 1)
}

func (rp *Replica) trigger() {
	for i := 0; ; i++ {
		if rp.State != 1 {
			break
		}

		fmt.Printf("trigger check:%v s\n", i)
		if (time.Now().UnixNano()-rp.ViewStartTime)/1e9 >= 4*config.D*int64(rp.View.ViewNum) { //timeout and trigger view-change
			fmt.Println("[State]: trigger view change")
			rp.State = 2
			rp.timeout()
			break
		}
		time.Sleep(time.Second)
	}
}

func (rp *Replica) timeout() {
	m, err := rp.timeoutMsg()
	if err != nil {
		return
	}
	fmt.Println("[State]: multicast a timeout msg of view", m.ViewNum)
	rp.ConnI.JsonSendToAll(*m, "/timeout")

	jsonMsg, _ := json.Marshal(*m)
	fmt.Printf("timeout size:%v bytes.\n", len(jsonMsg))

	rp.newView(m)
}

func (rp *Replica) timeoutMsg() (*TimeoutMsg, error) {
	var m *TimeoutMsg
	if rp.Voted.ViewNum == rp.View.ViewNum {
		fmt.Println("voted value")
		m = &TimeoutMsg{
			NodeID:     rp.node.NodeID,
			Value:      rp.Voted.Value,
			ViewNum:    rp.View.ViewNum,
			LeaderID:   rp.Voted.LeaderID,
			SignLeader: rp.Voted.SignLeader,
			SignIn:     rp.Voted.SignIn,
		}
	} else {
		fmt.Println("NO new voted value in view", rp.View.ViewNum)
		m = &TimeoutMsg{
			NodeID:     rp.node.NodeID,
			Value:      "",
			ViewNum:    rp.View.ViewNum,
			LeaderID:   "",
			SignLeader: crypto.ECDSAsign{},
		}
		m.SignIn = crypto.SignECDSA(m.Value+strconv.Itoa(rp.View.ViewNum), rp.node.NodeKey.NodeSk)
	}
	jsonMsg, err := json.Marshal(*m)
	if err != nil {
		fmt.Println("jsonMarshal err when creating timeout msg:", err)
		return &TimeoutMsg{}, err
	}
	m.SignOut = crypto.SignECDSA(string(jsonMsg), rp.node.NodeKey.NodeSk)

	return m, nil
}

func (rp *Replica) newView(msg *TimeoutMsg) {
	if rp.State == 1 {
		return
	}
	if rp.isValidTimeoutMsg(msg) && msg.ViewNum >= rp.View.ViewNum {
		rp.ValidTimeoutMsgs[msg.ViewNum] = append(rp.ValidTimeoutMsgs[msg.ViewNum], msg)
		fmt.Printf("[TIMEOUT]: receive a valid timeout of view %v from node %v, total %v\n", msg.ViewNum, msg.NodeID, len(rp.ValidTimeoutMsgs[msg.ViewNum]))

	} else {
		//fmt.Println("timeout check error")
		return
	}

	mms := rp.ValidTimeoutMsgs[msg.ViewNum]
	if len(mms) >= config.QC1 && isValidToGenTimeoutQC(mms) {
		mms = append(mms[:config.QC1-1], mms[len(mms)-1:]...)
		tQC := timeoutQCMsg(mms)
		lockType, v := LockOn(tQC)
		tQC.LockValue = v
		tQC.LockType = lockType
		fmt.Printf("[State]: generate a timeoutQC of view %v, %v timeoutMsg, locktype %v\n", msg.ViewNum, len(tQC.TQC), tQC.LockType)
		fmt.Println("multicast a timeoutQC")
		rp.ConnI.JsonSendToAll(*tQC, "/timeoutQC")
		delete(rp.ValidTimeoutMsgs, msg.ViewNum)
		rp.updateAndAdvance(tQC)
	}
}

func timeoutQCMsg(msgs []*TimeoutMsg) *TimeoutQC {
	mm := make([]TimeoutMsg, 0)
	for _, msg := range msgs {
		mm = append(mm, *msg)
	}
	m := &TimeoutQC{
		TQC:       mm,
		ViewNum:   mm[0].ViewNum,
		LockType:  0,
		LockValue: "",
	}
	return m
}

func (rp *Replica) recvTimeoutQC(msg *TimeoutQC) {
	if msg.ViewNum < rp.View.ViewNum {
		//fmt.Println("recv a timeoutQC, but i have been advanced")
		return
	}
	if rp.isValidTimeoutQC(msg) {
		fmt.Println("[State]: recv a valid timeoutQC and advance view")
		rp.updateAndAdvance(msg)
	}
}

func (rp *Replica) updateAndAdvance(msg *TimeoutQC) {
	if msg.LockValue != "" {
		fmt.Println("[State]: update highQC")
		rp.HighQC = msg
	}
	rp.View.ViewNum++
	rp.View.Leader = getViewLeader(rp.View.ViewNum)
	fmt.Printf("[State]: enter NEW view %v with leader of node %v\n", rp.View.ViewNum, rp.View.Leader)
	/*
		if rp.node.NodeID != rp.View.Leader {
			rp.StartTime = time.Now().UnixNano()
		}
	*/

	m, err := rp.statusMsg()
	if err != nil {
		return
	}
	if rp.node.NodeID != rp.View.Leader {
		rp.ConnI.JsonSendToLeader(*m, rp.node.NodeTable[rp.View.Leader], "/status")
		fmt.Println("[State]: send a status msg to node", rp.View.Leader)

		jsonMsg, _ := json.Marshal(*m)
		fmt.Printf("status msg size:%v bytes, highQC contains %v timeoutMsg of view %v.\n", len(jsonMsg), len(m.HighQC.TQC), m.HighQC.ViewNum)

		rp.State = 1
		fmt.Println()
		fmt.Println("[State]: Normal-case of view", rp.View.ViewNum)
		rp.ViewStartTime = time.Now().UnixNano()
		go rp.trigger()
	} else {
		rp.status(m)
	}
}

//statusMsg: a status msg contains the old view number.
func (rp *Replica) statusMsg() (*Status, error) {
	m := &Status{
		NodeID:  rp.node.NodeID,
		ViewNum: rp.View.ViewNum - 1,
		HighQC:  *rp.HighQC,
	}
	jsonMsg, err := json.Marshal(*m)
	if err != nil {
		return &Status{}, err
	}
	m.SignOut = crypto.SignECDSA(string(jsonMsg), rp.node.NodeKey.NodeSk)
	return m, nil
}

//status: for leader only
func (rp *Replica) status(msg *Status) {
	if rp.State != 2 {
		return
	}
	if rp.isValidStatusMsg(rp.View.ViewNum-1, msg) {
		fmt.Println("[STATUS MSG]: receive a valid status msg.")
		rp.ValidStatusMsgs = append(rp.ValidStatusMsgs, msg)
	}
	if len(rp.ValidStatusMsgs) >= config.QC2 {
		if rp.View.ViewNum < msg.ViewNum {
			rp.View.ViewNum++
			rp.View.Leader = getViewLeader(rp.View.ViewNum)
			fmt.Printf("[State]: enter NEW view %v (enough status) with leader of node %v\n", rp.View.ViewNum, rp.View.Leader)
		}

		proof := proofMsg(rp.ValidStatusMsgs, rp.View.ViewNum-1)

		rp.State = 1
		fmt.Println()
		fmt.Println("[State]: Normal-case of view", rp.View.ViewNum)
		rp.ViewStartTime = time.Now().UnixNano()
		go rp.trigger()
		rp.propose(proof)
	}
}

//proofMsg: each proof contains the new view number.
func proofMsg(msgs []*Status, w int) *Proof {
	hw := -1
	var v string // the value in the proof
	sc := make([]Status, 0)
	for _, st := range msgs {
		stw := st.HighQC.ViewNum
		if stw == w {
			sc = append(sc, *st)
			m := &Proof{
				Value:      st.HighQC.LockValue,
				ViewNum:    w + 1,
				StatusCert: sc,
			}
			return m
		}
		sc = append(sc, *st)

		if stw >= hw {
			hw = stw
			v = st.HighQC.LockValue
		}
	}

	m := &Proof{
		Value:      v,
		ViewNum:    w + 1,
		StatusCert: sc,
	}
	return m
}
