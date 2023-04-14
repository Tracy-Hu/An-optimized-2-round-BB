package consensus

import (
	"TwoRoundBB/crypto"
)

type ProposeMsg struct {
	NodeID  string           `json:"node_id"`
	Value   string           `json:"value"`
	ViewNum int              `json:"view_num"`
	SignIn  crypto.ECDSAsign `json:"sign_in"`
	Proof   Proof            `json:"proof"`
	SignOut crypto.ECDSAsign `json:"sign_out"`
}

type VoteMsg struct {
	NodeID     string           `json:"node_id"`
	Value      string           `json:"value"`
	ViewNum    int              `json:"view_num"`
	LeaderID   string           `json:"leader_id"`
	SignLeader crypto.ECDSAsign `json:"sign_leader"`
	SignIn     crypto.ECDSAsign `json:"sign_in"`
	SignOut    crypto.ECDSAsign `json:"sign_out"`
}

type TimeoutMsg struct {
	NodeID     string           `json:"node_id"`
	Value      string           `json:"value"` // "0": represents \bot value
	ViewNum    int              `json:"view_num"`
	LeaderID   string           `json:"leader_id"`
	SignLeader crypto.ECDSAsign `json:"sign_leader"`
	SignIn     crypto.ECDSAsign `json:"sign_in"`
	SignOut    crypto.ECDSAsign `json:"sign_out"`
}

// TimeoutQC LockType def:
// type 0: no lock on non-\bot value;
// type 1: contains at least 2F-1 valid timeoutMsg and no conflict value;
// type 2: contains at least 2F valid timeoutMsg except fo the leader.
type TimeoutQC struct {
	TQC       []TimeoutMsg `json:"tqc"`
	ViewNum   int          `json:"view_num"`
	LockType  int          `json:"lock_type"`
	LockValue string       `json:"lock_value"`
}

type Status struct {
	NodeID  string           `json:"node_id"`
	ViewNum int              `json:"view_num"`
	HighQC  TimeoutQC        `json:"high_qc"`
	SignOut crypto.ECDSAsign `json:"sign_out"`
}

type Proof struct {
	Value      string   `json:"value"`
	ViewNum    int      `json:"view_num"`
	StatusCert []Status `json:"status_cert"`
}

type CommitQC struct {
	NodeID string    `json:"node_id"`
	CQC    []VoteMsg `json:"cqc"`
}
