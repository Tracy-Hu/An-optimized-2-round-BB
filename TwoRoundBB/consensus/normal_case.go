package consensus

import (
	"TwoRoundBB/config"
	"TwoRoundBB/crypto"
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

func (rp *Replica) propose(proof *Proof) {
	/*
		//--- simulate faulty leaders ---//
		id, _ := strconv.Atoi(rp.node.NodeID)
		if id == 2 {
			fmt.Println("faulty leader refuses to propose.")
			return
		}
		//------------------------------//
	*/

	if rp.State != 1 {
		fmt.Println("[Fail]: state is not normal-case in propose...")
		return
	}
	fmt.Println("[State]: begin propose in view", rp.View.ViewNum)

	var v string
	prf := &Proof{}
	if rp.View.ViewNum == 1 {
		v = rp.node.InitValue
	} else {
		v = proof.Value
		if v == "" {
			v = rp.node.InitValue
		}
		prf = proof
	}
	m, err := rp.proposeMsg(v, prf)
	//fmt.Println("proof size:", len(m.Proof.StatusCert))
	//fmt.Println("proposal:", m)
	if err != nil {
		fmt.Println("[Fail]: fail to generate propose msg.")
		return
	}

	/*
		//--- simulate faulty leaders ---//
		if id == 1 {
			nodeid := make([]string, 0)
			nodeid = append(nodeid, strconv.Itoa(1))
			nodeid = append(nodeid, strconv.Itoa(2))
			fmt.Printf("propose to %v replicas.\n", len(nodeid))
			rp.ConnI.JsonSendToSome(*m, "/propose", nodeid)
		} else {
			rp.ConnI.JsonSendToAll(*m, "/propose")
		}
		//------------------------------//

	*/
	rp.ConnI.JsonSendToAll(*m, "/propose")

	jsonMsg1, _ := json.Marshal(m.Proof)
	fmt.Printf("proof size:%v bytes\n", len(jsonMsg1))

	jsonMsg, _ := json.Marshal(*m)
	fmt.Printf("proposal size:%v bytes, proof contains %v status\n", len(jsonMsg), len(m.Proof.StatusCert))

	if m.ViewNum != 1 {
		jsonMsg1, _ := json.Marshal(m.Proof.StatusCert[0])
		fmt.Printf("(per status %v bytes)\n", len(jsonMsg1))
	}

	rp.vote(m)
}

func (rp *Replica) proposeMsg(value string, proof *Proof) (*ProposeMsg, error) {
	m := &ProposeMsg{
		NodeID:  rp.node.NodeID,
		Value:   value,
		ViewNum: rp.View.ViewNum,
		SignIn:  crypto.SignECDSA(value+strconv.Itoa(rp.View.ViewNum), rp.node.NodeKey.NodeSk),
		Proof:   *proof,
	}
	jsonMsg, err := json.Marshal(*m)
	if err != nil {
		fmt.Println("[Error]: proposeMsg err:", err)
		return &ProposeMsg{}, nil
	}
	m.SignOut = crypto.SignECDSA(string(jsonMsg), rp.node.NodeKey.NodeSk)

	return m, nil
}

func (rp *Replica) vote(msg *ProposeMsg) {
	if rp.State != 1 {
		fmt.Println("[Fail]: state is not normal-case in vote....")
		return
	}

	if msg.ViewNum != rp.View.ViewNum {
		fmt.Printf("[Fail]: receive a inconsistent view's proposal (%v), my view %v.\n", msg.ViewNum, rp.View.ViewNum)
		return
	}

	b := true // msg contains a valid proof

	if !rp.isValidProposal(msg) {
		fmt.Println("[Fail]: invalid proposal.")
		return
	}
	if rp.View.ViewNum != 1 {
		b = rp.isValidProof(msg.Value, msg.ViewNum, &msg.Proof)
	}

	if b {
		fmt.Println("[PROPOSAL]: receive a valid proposal from node", msg.NodeID)
		m, err := rp.voteMsg(msg)
		if err != nil {
			fmt.Println("[Fail]: fail to generate a vote msg.")
			return
		}
		rp.Voted = m

		/*
			//--- simulate faulty leader 1, do not vote ---//
			if msg.NodeID == "1" {
				return
			}
			//--------------------------------------------//
		*/

		fmt.Println("begin vote")
		rp.ConnI.JsonSendToAll(*m, "/vote")

		jsonMsg, _ := json.Marshal(*m)
		fmt.Printf("vote size:%v bytes.\n", len(jsonMsg))

		rp.commit(m)
		fmt.Println("[State]: multicast a vote of view", msg.ViewNum)
	} else {
		fmt.Println("[Fail]: receive an invalid proposal")
	}
}

func (rp *Replica) voteMsg(msg *ProposeMsg) (*VoteMsg, error) {
	m := &VoteMsg{
		NodeID:     rp.node.NodeID,
		Value:      msg.Value,
		ViewNum:    msg.ViewNum,
		LeaderID:   rp.View.Leader,
		SignLeader: msg.SignIn,
		SignIn:     crypto.SignECDSA(msg.Value+strconv.Itoa(msg.ViewNum), rp.node.NodeKey.NodeSk),
	}

	jsonMsg, err := json.Marshal(*m)
	if err != nil {
		return &VoteMsg{}, err
	}
	m.SignOut = crypto.SignECDSA(string(jsonMsg), rp.node.NodeKey.NodeSk)

	if !crypto.VrfECDSA(rp.node.NodeKey.NodePks[m.LeaderID], m.SignLeader, m.Value+strconv.Itoa(m.ViewNum)) {
		fmt.Println("no")
	}
	return m, nil
}

func (rp *Replica) commit(msg *VoteMsg) {
	if rp.State == 3 {
		return
	}
	if rp.State != 1 {
		fmt.Println("[Fail]: state is not normal-case in commit....")
		return
	}

	if rp.isValidVote(msg) {
		rp.ValidVoteMsgs = append(rp.ValidVoteMsgs, msg)
		fmt.Println("[VOTE]: receive a valid vote, total", len(rp.ValidVoteMsgs))
	}
	if len(rp.ValidVoteMsgs) >= config.QC0 {
		b, cQC := rp.isValidCommitQC(rp.ValidVoteMsgs, false)
		if b {
			fmt.Printf("[COMMIT]: commit a value in view %v -- I'm Done!\n", rp.View.ViewNum)
			rp.State = 3
			//if rp.node.NodeID == rp.View.Leader {
			calLatency(rp.StartTime, time.Now().UnixNano())
			//}
			rp.ConnI.JsonSendToAll(*cQC, "/commitQC")
		}
	}
}

func (rp *Replica) decide(msg *CommitQC) {
	if rp.State == 3 {
		return
	}
	if rp.State != 1 {
		fmt.Println("[Fail]: state is not normal-case in decide....")
		return
	}

	rp.removeNode(msg.NodeID)
	m := make([]*VoteMsg, 0)
	for i := 0; i < len(msg.CQC); i++ {
		m = append(m, &msg.CQC[i])
	}

	b, _ := rp.isValidCommitQC(m, true)
	if b {
		fmt.Println("[COMMIT]: I'm Done with a commitQC.")
		calLatency(rp.StartTime, time.Now().UnixNano())
		rp.State = 3
		id, _ := strconv.Atoi(rp.node.NodeID)
		rp.ConnI.JsonSendToAll(id, "/done")
	}
}
