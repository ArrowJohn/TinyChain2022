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

	//calculate the nonce based on difficulty 0
	hex := fmt.Sprintf("%x", 0)
	for _, items := range General.QueryBlockChain() {
		POWTrans := transTransactionsToPOW(items.Transactions)
		blockBody := BlockBody{POWTrans}
		content := copyToContent(POWTrans)
		tr, _ := merkletree.NewTree(content)
		currentBlock := Block{
			BasicBlock: items,
			MerkleRoot: tr.MerkleRoot(),
			Difficulty: 1,
			Nonce:      hex,
			Body:       blockBody,
		}
		Blockchain = append(Blockchain, currentBlock)
	}

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
	_ = g.SendDirect(msg, BootNodeAddress)

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
