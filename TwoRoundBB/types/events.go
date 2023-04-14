package types

type ConsensusInterface interface {
	GetMsg(msg interface{})
}

type ConnInterface interface {
	JsonSendToLeader(msg interface{}, url string, path string)
	JsonSendToAll(msg interface{}, path string)
	JsonSendToSome(msg interface{}, path string, nodeid []string)
}
