package message

//JSONMessage : define the JSON formate message type
type JSONMessage struct {
	Type      string
	Body      string
	Time      string
	Signature string
	Round	  int
}

type MessageType struct {
	Bootstrap        string
	LatestBlockChain string
	CurrentWinner    string
	NewTransaction   string
	NewProposedBlock string
	NewJoinNode      string
	TEST             string
}

type MsgProtocol interface {
	Close() //to sefely terminates the messgaing protocl instance
}