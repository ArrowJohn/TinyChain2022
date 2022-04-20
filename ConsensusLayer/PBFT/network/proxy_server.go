package network

import (
	"TinyChain/ConsensusLayer/General"
	"TinyChain/ConsensusLayer/PBFT/consensus"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
	"time"
)

type Server struct {
	url  string
	node *Node
}

func NewServer(nodeID string, publicAddress string) *Server {
	node := NewNode(nodeID, publicAddress)
	server := &Server{node.NodeTable[nodeID], node}
	server.setRoute()
	return server
}

func (server *Server) setRoute() {
	r := mux.NewRouter()
	r.HandleFunc("/getBlockchainStatus", General.GetBlockchainStatus).Methods("GET")
	r.HandleFunc("/getBlockchain", General.GetBlockchain).Methods("GET")
	r.HandleFunc("/getPartBlockchain", General.GetPartBlockchain).Methods("GET")
	r.HandleFunc("/getAllTransactions", General.GetAllTransactions).Methods("GET")
	r.HandleFunc("/getTransaction/{address}", General.GetUserTransactions).Methods("GET")
	r.HandleFunc("/getLastestBlock", General.GetLatestBlock).Methods("GET")
	r.HandleFunc("/getBalance/{address}", General.GetBalance).Methods("GET")
	r.HandleFunc("/getTransactionByHash", General.GetTransByHash).Methods("GET")
	r.HandleFunc("/getBlockByIndex", General.GetBlockByIndex).Methods("GET")
	r.HandleFunc("/getCurrentTransactions", General.GetCurrentTrans).Methods("GET")
	r.HandleFunc("/sendTransaction", server.getReq).Methods("POST")
	r.HandleFunc("/preprepare", server.getPrePrepare)
	r.HandleFunc("/prepare", server.getPrepare)
	r.HandleFunc("/commit", server.getCommit)
	r.HandleFunc("/reply", server.getReply)
	println("Server run at " + server.url)
	log.Fatal(http.ListenAndServe(server.url, r))
}

func (server *Server) getReq(w http.ResponseWriter, request *http.Request) {
	var msg consensus.RequestMsg
	var trans General.Transaction
	_ = json.NewDecoder(request.Body).Decode(&trans)
	if !General.CheckTransValid(trans) {
		_ = json.NewEncoder(w).Encode(General.FormatResponse(500, "Something goes wrong please try again"))
		return
	}
	msg.Timestamp, _ = strconv.ParseInt(trans.Date, 10, 64)
	msg.ClientID = trans.From
	op, _ := json.Marshal(trans)
	msg.Operation = string(op)
	fmt.Println(msg)
	server.node.MsgEntrance <- &msg
	_ = json.NewEncoder(w).Encode(General.FormatResponse(200, "transaction created successfully"))
}

func (server *Server) getPrePrepare(_ http.ResponseWriter, request *http.Request) {
	var msg consensus.PrePrepareMsg
	err := json.NewDecoder(request.Body).Decode(&msg)
	if err != nil {
		fmt.Println(err)
		return
	}

	server.node.MsgEntrance <- &msg
}

func (server *Server) getPrepare(_ http.ResponseWriter, request *http.Request) {
	var msg consensus.VoteMsg
	err := json.NewDecoder(request.Body).Decode(&msg)
	if err != nil {
		fmt.Println(err)
		return
	}

	server.node.MsgEntrance <- &msg
}

func (server *Server) getCommit(_ http.ResponseWriter, request *http.Request) {
	var msg consensus.VoteMsg
	err := json.NewDecoder(request.Body).Decode(&msg)
	if err != nil {
		fmt.Println(err)
		return
	}

	server.node.MsgEntrance <- &msg
}

func (server *Server) getReply(_ http.ResponseWriter, request *http.Request) {
	var msg consensus.ReplyMsg
	err := json.NewDecoder(request.Body).Decode(&msg)
	if err != nil {
		fmt.Println(err)
		return
	}

	server.node.GetReply(&msg)
}

func send(url string, msg []byte) {
	buff := bytes.NewBuffer(msg)
	client := http.Client{Timeout: time.Second * 3}
	resp, err := client.Post("http://"+url, "application/json", buff)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	//_, err := http.Post("https://"+url, "application/json", buff)
}
