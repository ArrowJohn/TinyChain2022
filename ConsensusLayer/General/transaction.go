package General

import (
	_ "crypto/sha256"
	"fmt"
)

type Transaction struct {
	ID        int    `json:"id"`
	Date      string `json:"date"`
	From      string `json:"from"`
	To        string `json:"to"`
	Value     int    `json:"value"`
	Signature string `json:"signature"`
	Hash      string `json:"hash"`
}

func (transaction *Transaction) TableName() string {
	return "transactions"
}

func RemoveTransFromPool(trans []Transaction, transactions []Transaction) {
	var counter []int
	for i, storedTran := range transactions {
		for _, tran := range trans {
			if tran.Signature == storedTran.Signature {
				if !FindIndex(counter, i) {
					counter = append(counter, i)
				}
			}
		}
	}
	var newTrans []Transaction
	for i := range transactions {
		if !FindIndex(counter, i) {
			newTrans = append(newTrans, transactions[i])
		}
	}
	transactions = newTrans
}

func FindTrans(slice []Transaction, val Transaction) (int, bool) {
	for i, item := range slice {
		if item.Signature == val.Signature {
			return i, true
		}
	}
	return -1, false
}

// QueryTrans 查询全部TransList
func QueryTrans() []Transaction {
	var transList []Transaction
	GormDb.Find(&transList)
	return transList
}

func QueryTranByHash(hash string) Transaction {
	var transaction Transaction
	GormDb.First(&transaction, "hash=?", hash)
	return transaction
}

func QueryTranBySignature(signature string) Transaction {
	var transaction Transaction
	GormDb.First(&transaction, "signature=?", signature)
	return transaction
}

func InsertTransaction(transaction Transaction) {
	GormDb.Create(&transaction)
}

func InsertNewTransaction(transaction Transaction) {
	var trans Transaction
	GormDb.Last(&trans)
	transaction.ID = trans.ID + 1
	InsertTransaction(transaction)
}

func CheckTranExist(transaction Transaction) bool {
	println("check")
	transactions := QueryTrans()
	for _, trans := range transactions {
		if transaction.Hash == trans.Hash {
			return true
		}
	}
	return false
}

// InitTransaction 从库里面读取TransList, 没有的话就新建
func InitTransaction() []Transaction {
	var trans = QueryTrans()
	if len(trans) != 0 {
		fmt.Println("Read Transaction List")
		return trans
	}
	fmt.Println("Init Transaction list")
	timestamp := CurrentTimestamp()
	transaction1 := Transaction{
		Date: timestamp, From: "genesisTransaction", ID: 1,
		Signature: CalculateHash(timestamp + "genesisTransaction" + "1" + "YY" + string(rune(20))),
		To:        "db9cf6884b3983e488e4", Value: 1000,
	}
	transaction1.Hash = CalculateTranHash(transaction1)
	InsertTransaction(transaction1)
	transaction2 := Transaction{
		Date: timestamp, From: "genesisTransaction", ID: 2,
		Signature: CalculateHash(timestamp + "genesisTransaction" + "2" + "YY" + string(rune(20))),
		To:        "5c24967295a450bb96e3", Value: 1000,
	}
	transaction2.Hash = CalculateTranHash(transaction2)
	InsertTransaction(transaction2)
	transaction3 := Transaction{
		Date: timestamp, From: "genesisTransaction", ID: 3,
		Signature: CalculateHash(timestamp + "genesisTransaction" + "3" + "YY" + string(rune(20))),
		To:        "2975f996c46085cfecf6", Value: 1000,
	}
	transaction3.Hash = CalculateTranHash(transaction3)
	InsertTransaction(transaction3)
	InitBlock()
	return QueryTrans()
}

func GetBalanceFromTrans(address string) int {
	var outBalance int
	var inBalance int
	for _, tran := range QueryTrans() {
		if tran.From == address {
			outBalance += tran.Value

		}
		if tran.To == address {
			inBalance += tran.Value
		}
	}
	return inBalance - outBalance
}

func CheckTransValid(transaction Transaction) bool {
	if QueryTranBySignature(transaction.Signature).ID != 0 {
		return false
	}
	if GetBalanceFromTrans(transaction.From) < transaction.Value {
		return false
	}
	return true
}
