package consensus

import (
	"TwoRoundBB/config"
	"TwoRoundBB/crypto"
	"encoding/json"
	"fmt"
	"strconv"
)

func (rp *Replica) isValidProposal(msg *ProposeMsg) bool {
	if !crypto.VrfECDSA(rp.node.NodeKey.NodePks[msg.NodeID], msg.SignIn, msg.Value+strconv.Itoa(msg.ViewNum)) {
		fmt.Println("[Fail]: proposal signin fails.")
		return false
	}
	m := &ProposeMsg{
		NodeID:  msg.NodeID,
		Value:   msg.Value,
		ViewNum: msg.ViewNum,
		SignIn:  msg.SignIn,
		Proof:   msg.Proof,
	}
	jsonMsg, err := json.Marshal(*m)
	if err != nil {
		fmt.Println("[Error]: proposeMsg check err:", err)
		return false
	}
	if !crypto.VrfECDSA(rp.node.NodeKey.NodePks[msg.NodeID], msg.SignOut, string(jsonMsg)) {
		fmt.Println("[Fail]: proposal signout fails.")
		return false
	}
	return true
}

//isValidProof inputs value, view, proof
func (rp *Replica) isValidProof(v string, w int, proof *Proof) bool {
	if (proof.Value != "" && proof.Value != v) || proof.ViewNum != w {
		fmt.Println("[Fail]: proof value and view is not consistent with propose msg")
		return false
	}

	if len(proof.StatusCert) == 1 { // it contains only one timeoutQC of view w-1
		if proof.StatusCert[0].ViewNum != w-1 || proof.StatusCert[0].HighQC.ViewNum != w-1 {
			return false
		}
		if !rp.isValidStatusMsg(w-1, &proof.StatusCert[0]) {
			return false
		}
		if proof.StatusCert[0].HighQC.LockValue != v {
			fmt.Println("[Fail]: lock value in not the propose value.")
			return false
		}
	} else if len(proof.StatusCert) >= config.QC2 {
		var setTQC []*TimeoutQC
		for _, m := range proof.StatusCert {
			if !rp.isValidStatusMsg(w-1, &m) {
				fmt.Println("[Fail]: there is a invalid status msg in proof")
				return false
			}
			setTQC = append(setTQC, &m.HighQC)
		}
		hQC := highestQC(setTQC)
		if hQC.ViewNum > w-1 || hQC.LockValue != proof.Value {
			return false
		}

	} else {
		return false
	}

	return true
}

//highestQC inputs a set of valid timeoutQC, outputs the timeoutQC with the highest view number.
func highestQC(setTQC []*TimeoutQC) *TimeoutQC {
	w := 0
	t := 0
	for i, tQC := range setTQC {
		if tQC.ViewNum > w {
			t = i
			w = tQC.ViewNum
		}
	}
	return setTQC[t]
}

//isValidStatusMsg checks: view number, signature, highQC (validity of lock value).
func (rp *Replica) isValidStatusMsg(w int, msg *Status) bool {
	if msg.ViewNum < w {
		fmt.Println("[Fail]: view number check fail in isValidStatusMsg.")
		fmt.Printf("view number:%v, status view number:%v.\n", w, msg.ViewNum)
		return false
	}
	//check status msg signature
	mm := &Status{
		NodeID:  msg.NodeID,
		ViewNum: msg.ViewNum,
		HighQC:  msg.HighQC,
	}
	jsonMsg, err := json.Marshal(*mm)
	if err != nil {
		fmt.Println("[Error]: isValidStatusMsg error:", err)
		return false
	}

	if !crypto.VrfECDSA(rp.node.NodeKey.NodePks[msg.NodeID], msg.SignOut, string(jsonMsg)) {
		fmt.Println("[Fail]: isValidStatusMsg signature fail.")
		return false
	}

	//check highQC
	highQC := msg.HighQC
	b := rp.isValidTimeoutQC(&highQC)
	if !b || highQC.ViewNum > msg.ViewNum {
		if !b {
			fmt.Println("isValidTimeoutQC check fail")
		} else {
			fmt.Println("highQC viewnumber check fail:", highQC.ViewNum)
		}
		fmt.Printf("[Fail]: isValidTimeoutQC from node %v check fails in isValidStatusMsg\n", msg.NodeID)
		return false
	}
	return true
}

// isValidToGenTimeoutQC follows the protocol rule
func isValidToGenTimeoutQC(msgs []*TimeoutMsg) bool {
	// check # of no-\bot values in msgs
	v := make([]string, 0)
	for _, msg := range msgs {
		if msg.Value != "" {
			v = append(v, msg.Value)
		}
	}
	if len(v) == 1 {
		return true
	} else if len(v) == 2 && (v[0] == "" || v[1] == "") {
		return true
	} else {
		for i, msg := range msgs {
			if msg.NodeID == msg.LeaderID {
				msgs = append(msgs[:i], msgs[i+1:]...) // remove the leader's timeout msg
				break
			}
		}
		if len(msgs) >= config.QC1 {
			return true
		}
	}
	return false
}

//isValidTimeoutQC checks: number of timeout msgs, timeout msg signatures, repetitive nodes, lock value validity.
func (rp *Replica) isValidTimeoutQC(tQC *TimeoutQC) bool {
	if tQC.ViewNum == -1 {
		return true
	}

	if len(tQC.TQC) < config.QC1 {
		fmt.Println("[Fail]: isValidTimeoutQC verify fail: not enough msgs.")
		return false
	}
	for _, m := range tQC.TQC {
		if !rp.isValidTimeoutMsg(&m) || m.ViewNum != tQC.ViewNum {
			fmt.Println("[Fail]: isValidTimeoutQC verify fail.")
			return false
		}
	}

	//check for repetitive timeout msgs
	nodeids := make([]string, 0)
	for _, m := range tQC.TQC {
		for _, id := range nodeids {
			if m.NodeID == id {
				fmt.Println("[Fail]: repetitive timeout msgs in isValidTimeoutQC.")
				return false
			}
		}
		nodeids = append(nodeids, m.NodeID)
	}

	//check for valid lock value
	lockType, v := LockOn(tQC)
	if v != tQC.LockValue || lockType != tQC.LockType {
		fmt.Println("[Faile]: no lock check in isValidTimeoutQC.")
		return false
	}

	return true
}

// LockOn return: lock type, the value that lock on
// type 0: no lock on non-\bot value;
// type 1: contains at least 2F-1 valid timeoutMsg and no conflict value;
// type 2: contains at least 2F valid timeoutMsg except fo the leader.
func LockOn(msg *TimeoutQC) (int, string) {
	msgs := msg.TQC
	vs := make([]string, 0)
	vs = append(vs, msgs[0].Value)

	// check how many values exist
	for _, tmsg := range msgs {
		for _, v := range vs {
			if tmsg.Value != v {
				vs = append(vs, tmsg.Value)
			}
		}
	}

	if len(vs) == 1 {
		if vs[0] == "" {
			//fmt.Println("empty lock on")
			return 0, ""
		}
		//fmt.Println("lock on, case 1")
		return 1, vs[0]
	} else if len(vs) == 2 && (vs[0] == "" || vs[1] == "") {
		var v string
		counter := 0
		for _, tmsg := range msgs {
			if tmsg.Value != "" {
				counter++
				v = tmsg.Value
			}
		}
		if counter >= 2*config.F-1 {
			//fmt.Println("lock on, case 1")
			return 1, v
		} else {
			//fmt.Println("no lock on, no case 1")
			return 0, ""
		}
	} else { // multiple values
		SS := make(map[string][]*TimeoutMsg)
		for _, tmsg := range msgs {
			if tmsg.NodeID == tmsg.LeaderID {
				continue
			}
			for _, v := range vs {
				if tmsg.Value == v {
					SS[v] = append(SS[v], &tmsg)
				}
			}
		}
		for _, ss := range SS {
			if len(ss) >= 2*config.F {
				//fmt.Println("lock on, case 2")
				return 2, ss[0].Value
			}
		}
		//fmt.Println("false lock on")
		return 0, ""
	}
}

// isValidTimeoutMsg checks: signature
func (rp *Replica) isValidTimeoutMsg(msg *TimeoutMsg) bool {
	m := &TimeoutMsg{
		NodeID:     msg.NodeID,
		Value:      msg.Value,
		ViewNum:    msg.ViewNum,
		LeaderID:   msg.LeaderID,
		SignLeader: msg.SignLeader,
		SignIn:     msg.SignIn,
	}
	jsonMsg, err := json.Marshal(*m)
	if err != nil {
		return false
	}

	if crypto.VrfECDSA(rp.node.NodeKey.NodePks[msg.NodeID], msg.SignOut, string(jsonMsg)) &&
		crypto.VrfECDSA(rp.node.NodeKey.NodePks[msg.NodeID], msg.SignIn, msg.Value+strconv.Itoa(msg.ViewNum)) {
		if msg.Value == "" {
			return true
		} else if crypto.VrfECDSA(rp.node.NodeKey.NodePks[msg.LeaderID], msg.SignLeader, msg.Value+strconv.Itoa(msg.ViewNum)) {
			return true
		}
	}

	return false
}

//isValidVote checks: signatures in a vote as an input.
func (rp *Replica) isValidVote(msg *VoteMsg) bool {
	if !crypto.VrfECDSA(rp.node.NodeKey.NodePks[msg.LeaderID], msg.SignLeader, msg.Value+strconv.Itoa(msg.ViewNum)) {
		fmt.Println("[Fail]: verify vote leader sign fail.")
		return false
	} else if !crypto.VrfECDSA(rp.node.NodeKey.NodePks[msg.NodeID], msg.SignIn, msg.Value+strconv.Itoa(msg.ViewNum)) {
		fmt.Println("[Fail]: verify vote sign_in fail.")
		return false
	} else {
		m := VoteMsg{
			NodeID:     msg.NodeID,
			Value:      msg.Value,
			ViewNum:    msg.ViewNum,
			LeaderID:   msg.LeaderID,
			SignLeader: msg.SignLeader,
			SignIn:     msg.SignIn,
		}

		jsonMsg, err := json.Marshal(m)
		if err != nil {
			fmt.Println("[Error]: isValidVote json error.")
			return false
		}
		if !crypto.VrfECDSA(rp.node.NodeKey.NodePks[msg.NodeID], msg.SignOut, string(jsonMsg)) {
			fmt.Println("[Fail]: verify vote sign_out fail.")
			return false
		}
	}
	return true
}

//isValidCommitQC checks: number of votes, repetitive votes.
//input: if b=true, verify each vote msgs; therwise, no need to verify each votemsg.
func (rp *Replica) isValidCommitQC(votemsgs []*VoteMsg, b bool) (bool, *CommitQC) {
	if b {
		for _, m := range votemsgs {
			if !rp.isValidVote(m) {
				return false, &CommitQC{}
			}
		}
	}

	w := votemsgs[0].ViewNum
	v := votemsgs[0].Value
	cQC := make([]VoteMsg, 0)
	nodeids := make([]string, 0)
	for _, m := range votemsgs {
		if m.ViewNum == w || m.Value == v {
			cQC = append(cQC, *m)
			for _, id := range nodeids {
				if m.NodeID == id {
					fmt.Println("[Fail]: repetitive vote msgs.")
					return false, &CommitQC{}
				}
			}
		}
	}
	if len(cQC) >= config.QC0 {
		m := &CommitQC{rp.node.NodeID, cQC}
		return true, m
	}
	return false, &CommitQC{}
}

func (rp *Replica) removeNode(nodeid string) {
	//fmt.Println("[REMOVE]: node", nodeid)
	delete(rp.node.NodeTable, nodeid)
}

func calLatency(tStart, tEnd int64) {
	fmt.Printf("Latency is: %v ms\n", (tEnd-tStart)/1e6)
}
