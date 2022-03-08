package main

import (
	"TinyChain/ConsensusLayer/PBFT/network"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

var transIndex int
var Addresses []string
var BootNodeAddress string

//PublicAdd public cloud address
var PublicAdd string

// PrivateAdd private cloud address
var PrivateAdd string

// Port for gp
var Port string

var ServerPort string

type Address struct {
	Address string `json:"address"`
}

type Config struct {
	PublicAdd        string    `json:"PublicAdd"`
	PrivateAdd       string    `json:"PrivateAdd"`
	Port             string    `json:"Port"`
	ServerPort       string    `json:"ServerPort"`
	Strategy         string    `json:"Strategy"`
	Addresses        []Address `json:"addresses"`
	TransactionIndex int       `json:"transIndex"`
}

type PBTF struct {
}

func (pbtf *PBTF) GetAllTransactions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("content-type", "application/json")
}

func (pbtf *PBTF) GetUserTransactions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("content-type", "application/json")
}

func (pbtf *PBTF) GetLatestBlock(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("content-type", "application/json")

}

func (pbtf *PBTF) GetBlockchain(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("content-type", "application/json")

}

func (pbtf *PBTF) GetPartBlockchain(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("content-type", "application/json")
}

func (pbtf *PBTF) GetBlockchainStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("content-type", "application/json")
}

func (pbtf *PBTF) GetBalance(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("content-type", "application/json")
}

func (pbtf *PBTF) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("content-type", "application/json")

}

func readConfig() {
	jsonFile, err := os.Open("config.json")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Successfully Opened users.json")
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
	} else {
		BootNodeAddress = config.Strategy
	}
	PublicAdd = config.PublicAdd
	PrivateAdd = config.PrivateAdd
	Port = config.Port
	ServerPort = config.ServerPort
}
func main() {
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

	nodeID := os.Args[1]
	server := network.NewServer(nodeID)
	server.Start()
	log.Fatal(http.ListenAndServe(":"+ServerPort, r))
}
