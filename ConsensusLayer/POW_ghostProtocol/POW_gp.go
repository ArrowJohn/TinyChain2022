package main

import (
	"TinyChain/ConsensusLayer/General"
	"TinyChain/Network/gossip"
	"TinyChain/Network/message"
	"encoding/json"
	"fmt"
	"github.com/cbergoon/merkletree"
	"github.com/davecgh/go-spew/spew"
	"log"
	"strings"
	"time"
)

func main() {
	readConfig()
	General.Connect()
	General.InitTransaction()
	timestamp := General.CurrentTimestamp()
	genesisBlock := Block{}
	genesisBlockBody := BlockBody{}
	transaction1 := POWTransaction{
		Transaction: General.Transaction{
			Date: timestamp, From: "genesisTransaction", ID: "0",
			Signature: General.CalculateHash(timestamp + "genesisTransaction" + "0" + "YY" + string(rune(20))),
			To:        Addresses[0], Value: 1000,
		},
	}
	transaction1.Hash = General.CalculateTranHash(unifyTransaction(transaction1))
	transaction2 := POWTransaction{
		Transaction: General.Transaction{
			Date: timestamp, From: "genesisTransaction", ID: "1",
			Signature: General.CalculateHash(timestamp + "genesisTransaction" + "1" + "YY" + string(rune(20))),
			To:        Addresses[1], Value: 2000,
		},
	}
	transaction2.Hash = General.CalculateTranHash(unifyTransaction(transaction2))
	//declare new transactions list and add the genesis trans in
	var genesisTransactions []POWTransaction
	genesisTransactions = append(genesisTransactions, transaction1)
	genesisTransactions = append(genesisTransactions, transaction2)

	//convert into content list so that we can use the merkle tree package
	var genesisTransactionsContent = copyToContent(genesisTransactions)
	tr, err := merkletree.NewTree(genesisTransactionsContent)
	if err != nil {
		log.Fatal(err)
	}

	//BlockBody
	genesisBlockBody = BlockBody{genesisTransactions}

	//calculate the nonce based on difficulty 0
	hex := fmt.Sprintf("%x", 0)
	genesisBlock = Block{
		BasicBlock: General.BasicBlock{Timestamp: General.CurrentTimestamp(), Hash: calculateBlockHash(genesisBlock), Signature: General.CalculateHash("genesis")},
		MerkleRoot: tr.MerkleRoot(), Difficulty: 1, Nonce: hex, Body: genesisBlockBody}

	spew.Dump(genesisBlock)
	Blockchain = append(Blockchain, genesisBlock)

	g = gossip.NewGossip(PublicAdd, PrivateAdd, Port, 3)
	go g.ReceiveLoop()

	go func() {
		for {
			select {
			case msg := <-g.NewMsg:
				msgHandler(msg)
			}
		}
	}()

	//choose the validator
	go func() {
		for {
			pickWinner()
		}
	}()

	userInput := make(chan string)

	go General.ReadInput(userInput)

	msg := message.JSONMessage{Type: messageType.Bootstrap, Body: PublicAdd + ":" + Port, Time: time.Now().String(), Signature: address}
	//send address to the bootnode
	err = g.SendDirect(msg, BootNodeAddress)
	if err != nil {
		return
	}

	go func() {
		for {
			select {
			case strs := <-userInput:
				if strings.HasPrefix("boot", strs) {
					fmt.Println("Sending message to BootNode")
					msg := message.JSONMessage{Type: messageType.Bootstrap, Body: PublicAdd + ":" + Port, Time: time.Now().String(), Signature: address}
					//send address to the bootnode
					//g.SendDirect(msg, "13.211.132.135:3333")
					err := g.SendDirect(msg, BootNodeAddress)
					if err != nil {
						return
					}
				}
			}
		}
	}()

	go func() {
		for {
			//periodically broadcast the current blockchain
			time.Sleep(10 * time.Second)
			output, err := json.Marshal(Blockchain)
			if err != nil {
				log.Fatal(err)
			}
			stdout <- Blockchain
			broadcastBlockchain(string(output), messageType.LatestBlockChain)
		}
	}()

	//general print function for the system
	go func() {
		for {
			select {
			case str := <-stdout:
				spew.Dump(str)
				fmt.Println("================================this is: " + PublicAdd + ":" + Port + "=======================================")
			}
		}
	}()

	setRouter()

}
