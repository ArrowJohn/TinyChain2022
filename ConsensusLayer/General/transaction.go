package General

import (
	_ "crypto/sha256"
)

type Transaction struct {
	ID        string `json:"id"`
	Date      string `json:"date"`
	From      string `json:"from"`
	To        string `json:"to"`
	Value     int    `json:"value"`
	Signature string `json:"signature"`
	Hash      string `json:"hash"`
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

func FindTransByHash(slice []Transaction, hash string) Transaction {
	for _, item := range slice {
		if item.Hash == hash {
			return item
		}
	}
	return Transaction{}
}
