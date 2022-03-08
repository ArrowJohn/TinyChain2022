package main

import (
	"TinyChain/Network/gossip"
	"TinyChain/Network/message"
	"bufio"
	"fmt"
	"os"
	"time"
)

const (
	//CONN_HOST = "172.31.32.105"
	CONN_HOST = "localhost"
	CONN_PORT = "8001"
	CONN_TYPE = "tcp"
)

type Node struct {
	content string
}

func testHandler(msg message.JSONMessage, n *Node) {
	n.content = msg.Body
	fmt.Printf(" -> TEST Node A read message: %s\n", msg.Body)
}

//example of how the message can update field in node
func getMsg(g *gossip.Gossip, n *Node) {
	for {
		select {
		case msg := <-g.NewMsg:
			testHandler(msg, n)
		}
	}
}

var messageType = message.MessageType{Bootstrap: "NewJoinNode", LatestBlockChain: "Bootstrap", CurrentWinner: "TEST", NewTransaction: "CurrentWinner", NewProposedBlock: "NewTransaction", NewJoinNode: "NewProposedBlock", TEST: "LatestBlockChain"}

func main() {

	g := gossip.NewGossip(CONN_HOST, CONN_HOST, CONN_PORT, 2)

	n := Node{content: "init"}

	g.AddPeer(CONN_HOST + ":8002")
	// g.AddPeer("localhost:8003")
	// g.AddPeer("localhost:8004")
	// g.AddPeer("localhost:8005")

	go g.ReceiveLoop()
	go getMsg(g, &n)

	for {
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		if input == "boot\n" {
			fmt.Println("BOOT")
			msg := message.JSONMessage{Type: messageType.Bootstrap, Body: CONN_HOST + ":8080", Time: time.Now().String(), Signature: "signature"}
			g.SendDirect(msg,  CONN_HOST + ":3333")
		} else {
			msg := message.JSONMessage{Type: "TEST", Body: input, Time: time.Now().String(), Signature: "signature"}
			g.Broadcast(msg)
		}
	}
}
