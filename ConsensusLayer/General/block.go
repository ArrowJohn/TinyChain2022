package General

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
)

type BasicBlock struct {
	Index        int
	Timestamp    string `gorm:"column:timeStamp"`
	Hash         string
	PrevHash     string
	Signature    string
	Transactions TransactionList
}

type TransactionList []Transaction

func (basicBlock *BasicBlock) TableName() string {
	return "blockChain"
}

func (transactionList *TransactionList) Scan(value interface{}) error {
	bytesValue, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal TransactionList value:", value))
	}
	return json.Unmarshal(bytesValue, transactionList)
}

func (transactionList TransactionList) Value() (driver.Value, error) {
	result, _ := json.Marshal(transactionList)
	return string(result), nil
}

func CalculateBlockHash(block BasicBlock) string {
	record := string(rune(block.Index)) + block.Timestamp + block.PrevHash + block.Signature
	return CalculateHash(record)
}

func QueryBlockChain() []BasicBlock {
	var blockChain []BasicBlock
	GormDb.Find(&blockChain)
	return blockChain
}

func QueryLastestBlock() BasicBlock {
	var block BasicBlock
	GormDb.Last(&block)
	return block
}

func QueryBlockByIndex(index int) BasicBlock {
	var basicBlock BasicBlock
	GormDb.First(&basicBlock, index)
	return basicBlock
}

func GetStatus() Detail {
	detail := Detail{
		BlockChainHeight:  len(QueryBlockChain()),
		TransactionNumber: len(QueryTrans()),
	}
	return detail
}

func QueryTransInBlockByIndex(index int) []Transaction {
	var transactions string
	var transactionList []Transaction
	blockChain := QueryBlockChain()
	for _, basicBlock := range blockChain {
		if basicBlock.Index == index {
			listStrByte := []byte(transactions)
			_ = json.Unmarshal(listStrByte, &transactionList)
		}
	}
	return transactionList
}

func QueryTransInBlock() []Transaction {
	var transactionList []Transaction
	blockChain := QueryBlockChain()
	for _, basicBlock := range blockChain {
		for _, transactions := range basicBlock.Transactions {
			transactionList = append(transactionList, transactions)
		}
	}
	return transactionList
}

func CheckTransInBlock(transaction Transaction) bool {
	result, _ := json.Marshal(transaction)
	transactionList := QueryTransInBlock()
	for _, trans := range transactionList {
		result1, _ := json.Marshal(trans)
		if string(result) == string(result1) {
			return true
		}
	}
	return false
}

func InsertBlock(block BasicBlock) {
	GormDb.Create(&block)
}

func InsertNewBlock(trans []Transaction) BasicBlock {
	lastBlock := QueryLastestBlock()
	var block BasicBlock
	block.Index = lastBlock.Index + 1
	block.Timestamp = CurrentTimestamp()
	block.PrevHash = lastBlock.Hash
	block.Signature = CalculateHash(block.Timestamp + block.PrevHash)
	block.Transactions = trans
	block.Hash = CalculateBlockHash(block)
	InsertBlock(block)
	return block
}

// InitBlock 根据 trans 初始化 block
func InitBlock() {
	ClearTable("blockchain")
	fmt.Println("Init TinyChain")
	timestamp := CurrentTimestamp()
	var trans = QueryTrans()
	block := BasicBlock{
		Index:     0,
		Timestamp: timestamp,
		Hash:      "",
		PrevHash:  "",
		Signature: CalculateHash("genesis"),
	}
	block.Transactions = trans
	block.Hash = CalculateBlockHash(block)
	//transStr, _ := json.Marshal(trans)
	//block.Transactions = string(transStr)
	InsertBlock(block)
}
