package General

import (
	"encoding/json"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
)

type BlockBody struct {
	Transactions []Transaction
}

type BasicBlock struct {
	Index        int
	Timestamp    string
	Hash         string
	PrevHash     string
	Signature    string
	Transactions string
}

func (basicBlock *BasicBlock) TableName() string {
	return "chain"
}

func CalculateBlockHash(block BasicBlock) string {
	record := string(rune(block.Index)) + block.Timestamp + block.PrevHash + block.Signature
	return CalculateHash(record)
}

func QueryBlocks() []BasicBlock {
	var blockChain []BasicBlock
	var trans string
	rows, e := db.Query("select * from chain")
	if e == nil {
		errors.New("query incur error")
	}
	for rows.Next() {
		var block BasicBlock
		_ = rows.Scan(
			&block.Index, &block.Hash, &block.PrevHash, &block.Timestamp, &block.Signature, &trans)
		blockChain = append(blockChain, block)
	}
	fmt.Println(blockChain)
	return blockChain
}

func QueryTransInBlock(index int) []Transaction {
	var transactions string
	var transactionList []Transaction
	rows, e := db.Query("select * from chain")
	if e == nil {
		errors.New("query incur error")
	}
	for rows.Next() {
		var block BasicBlock
		_ = rows.Scan(
			&block.Index, &block.Hash, &block.PrevHash, &block.Timestamp, &block.Signature, &transactions)
		if block.Index == index {
			listStrByte := []byte(transactions)
			_ = json.Unmarshal(listStrByte, &transactionList)
		}
	}
	return transactionList
}

func InsertBlock(block BasicBlock) {
	GormDb.Create(&block)
}

// InitBlock 根据 trans 初始化 block
func InitBlock() {
	ClearTable("chain")
	fmt.Println("Init TinyChain")
	timestamp := CurrentTimestamp()
	var trans = QueryTran()
	block := BasicBlock{
		Index:     0,
		Timestamp: timestamp,
		Hash:      "",
		PrevHash:  "",
		Signature: CalculateHash("genesis"),
	}
	block.Hash = CalculateBlockHash(block)
	transStr, _ := json.Marshal(trans)
	block.Transactions = string(transStr)
	InsertBlock(block)
}
