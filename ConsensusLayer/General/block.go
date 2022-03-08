package General

type BlockBody struct {
	Transactions []Transaction
}

type BasicBlock struct {
	Index     int
	Timestamp string
	Hash      string
	PrevHash  string
	Signature string
}
