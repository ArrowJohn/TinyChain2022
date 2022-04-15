package main

import (
	"TinyChain/ConsensusLayer/General"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

func readConfig() {
	jsonFile, err := os.Open("PBFT.json")
	if err != nil {
		jsonFile, _ = os.Open("ConsensusLayer/PBFT/PBFT.json")
	}
	byteValue, _ := ioutil.ReadAll(jsonFile)

	var config Config
	json.Unmarshal(byteValue, &config)

	for i := 0; i < len(config.Addresses); i++ {
		Addresses = append(Addresses, config.Addresses[i].Address)
	}
	transIndex = config.TransactionIndex

	fmt.Println(config)
	if config.Strategy == "local" {
		BootNodeAddress = "localhost:3333"
		PublicAdd = "localhost"
		PrivateAdd = "localhost"
	} else {
		BootNodeAddress = config.Strategy
		PublicAdd = config.PublicAdd
		PrivateAdd = config.PrivateAdd
	}
	Port = config.Port
	ServerPort = config.ServerPort
}

func calculateBlockHash(block Block) string {
	record := string(rune(block.Index)) + block.Timestamp + block.PrevHash +
		General.CalculateTransHash(unifyTransactions(block.Body.Transactions))
	return General.CalculateHash(record)
}

func unifyTransactions(transactions []PBFTTransaction) []General.Transaction {
	var transactionsGeneral []General.Transaction
	for _, item := range transactions {
		transactionsGeneral = append(transactionsGeneral, unifyTransaction(item))
	}
	return transactionsGeneral
}

func unifyTransaction(transaction PBFTTransaction) General.Transaction {
	transactionGeneral := General.Transaction{}
	data, _ := json.Marshal(transaction)
	_ = json.Unmarshal(data, &transactionGeneral)
	return transactionGeneral
}

func (pow *PBTF) GetAllTransactions(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("content-type", "application/json")
}

func (pow *PBTF) GetUserTransactions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("content-type", "application/json")
}

func (pow *PBTF) GetLatestBlock(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("content-type", "application/json")
}

func (pow *PBTF) GetBlockchain(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("content-type", "application/json")
}

func (pow *PBTF) GetPartBlockchain(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("content-type", "application/json")
}

func (pow *PBTF) GetBlockByIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("content-type", "application/json")
}

func (pow *PBTF) GetBlockchainStatus(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("content-type", "application/json")
}

func (pow *PBTF) GetBalance(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("content-type", "application/json")
}

func (pow *PBTF) GetTransByHash(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("content-type", "application/json")
}

func (pow *PBTF) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("content-type", "application/json")
}

func (pow *PBTF) GetCurrentTrans(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("content-type", "application/json")
}

func setRouter() {
	r := mux.NewRouter()
	pbtf := PBTF{}
	r.HandleFunc("/getBlockchainStatus", pbtf.GetBlockchainStatus).Methods("GET")
	r.HandleFunc("/getBlockchain", pbtf.GetBlockchain).Methods("GET")
	r.HandleFunc("/getPartBlockchain", pbtf.GetPartBlockchain).Methods("GET")
	r.HandleFunc("/getTransaction/{address}", pbtf.GetUserTransactions).Methods("GET")
	r.HandleFunc("/sendTransaction", pbtf.CreateTransaction).Methods("POST")
	r.HandleFunc("/getLastestBlock", pbtf.GetLatestBlock).Methods("GET")
	r.HandleFunc("/getBalance/{address}", pbtf.GetBalance).Methods("GET")
	r.HandleFunc("/getAllTransactions", pbtf.GetAllTransactions).Methods("GET")
	r.HandleFunc("/getTransactionByHash", pbtf.GetTransByHash).Methods("GET")
	r.HandleFunc("/getBlockByIndex", pbtf.GetBlockByIndex).Methods("GET")
	r.HandleFunc("/getCurrentTransactions", pbtf.GetCurrentTrans).Methods("GET")

	log.Fatal(http.ListenAndServe(":"+ServerPort, r))
}
