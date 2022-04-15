package main

import (
	"TinyChain/ConsensusLayer/General"
	"TinyChain/Network/gossip"
	"github.com/davecgh/go-spew/spew"
)

func main() {
	readConfig()
	timestamp := General.CurrentTimestamp()
	genesisBlock := Block{}
	genesisBlockBody := BlockBody{}
	transaction1 := PBFTTransaction{
		Transaction: General.Transaction{
			Date: timestamp, From: "genesisTransaction", ID: "0",
			Signature: General.CalculateHash(timestamp + "genesisTransaction" + "0" + "YY" + string(rune(20))),
			To:        Addresses[0], Value: 1000,
		},
	}
	transaction1.Hash = General.CalculateTranHash(unifyTransaction(transaction1))
	transaction2 := PBFTTransaction{
		Transaction: General.Transaction{
			Date: timestamp, From: "genesisTransaction", ID: "1",
			Signature: General.CalculateHash(timestamp + "genesisTransaction" + "1" + "YY" + string(rune(20))),
			To:        Addresses[1], Value: 2000,
		},
	}
	transaction2.Hash = General.CalculateTranHash(unifyTransaction(transaction2))

	var genesisTransactions []PBFTTransaction
	genesisTransactions = append(genesisTransactions, transaction1)
	genesisTransactions = append(genesisTransactions, transaction2)

	genesisBlockBody = BlockBody{genesisTransactions}
	genesisBlock = Block{
		BasicBlock: General.BasicBlock{Timestamp: General.CurrentTimestamp(), Hash: calculateBlockHash(genesisBlock), Signature: General.CalculateHash("genesis")},
		Body:       genesisBlockBody,
	}

	spew.Dump(genesisBlock)
	Blockchain = append(Blockchain, genesisBlock)

	g = gossip.NewGossip(PublicAdd, PrivateAdd, Port, 3)
	go g.ReceiveLoop()

	setRouter()

}
