package General

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

// Detail struct for some basic information of the chain
type Detail struct {
	BlockChainHeight  int `json:"blockChainHeight"`
	TransactionNumber int `json:"TransactionNumber"`
}

// Balance struct is a struct which stores the balance of the user address
type Balance struct {
	Address string `json:"address"`
	Balance int    `json:"balance"`
}

func GetBlockchain(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("content-type", "application/json")
	_ = json.NewEncoder(w).Encode(QueryBlockChain())
}

func GetBlockchainStatus(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("content-type", "application/json")
	_ = json.NewEncoder(w).Encode(GetStatus())
}

func GetPartBlockchain(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("content-type", "application/json")
	query := r.URL.Query()
	Blockchain := QueryBlockChain()
	pageIndex, _ := strconv.Atoi(query.Get("Index"))
	pageSize, _ := strconv.Atoi(query.Get("Size"))
	if pageIndex < 1 || (pageIndex-1)*pageSize >= len(Blockchain) {
		err := json.NewEncoder(w).Encode(FormatResponse(500, "Something goes wrong please try again"))
		if err != nil {
			return
		}
		return
	}
	var BlockchainPart []BasicBlock
	for i := (pageIndex - 1) * pageSize; i < pageIndex*pageSize; i++ {
		if i > len(Blockchain)-1 {
			break
		}
		BlockchainPart = append(BlockchainPart, Blockchain[i])
	}
	err := json.NewEncoder(w).Encode(BlockchainPart)
	if err != nil {
		return
	}
}

func GetUserTransactions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("content-type", "application/json")
	vars := mux.Vars(r)
	key := vars["address"]
	var trans []Transaction
	for _, block := range QueryBlockChain() {
		for _, tran := range block.Transactions {
			if tran.From == key || tran.To == key {
				trans = append(trans, tran)
			}
		}
	}
	err := json.NewEncoder(w).Encode(trans)
	if err != nil {
		return
	}
}

func GetAllTransactions(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("content-type", "application/json")
	_ = json.NewEncoder(w).Encode(QueryTrans())
}

func GetLatestBlock(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("content-type", "application/json")
	Blockchain := QueryBlockChain()
	err := json.NewEncoder(w).Encode(Blockchain[len(Blockchain)-1])
	if err != nil {
		return
	}
}

func GetBlockByIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("content-type", "application/json")
	query := r.URL.Query()
	index, _ := strconv.Atoi(query.Get("index"))
	block := QueryBlockByIndex(index)
	if block.Index != 0 {
		_ = json.NewEncoder(w).Encode(block)
		return
	} else {
		_ = json.NewEncoder(w).Encode(FormatResponse(500, "No such block"))
	}
}

func GetCurrentTrans(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("content-type", "application/json")
	query := r.URL.Query()
	user := query.Get("address")
	days, _ := strconv.Atoi(query.Get("days"))
	var trans []Transaction
	for _, block := range QueryBlockChain() {
		for _, tran := range block.Transactions {
			if (tran.From == user || tran.To == user) && JudgeInDays(tran.Date, days) {
				trans = append(trans, tran)
			}
		}
	}
	err := json.NewEncoder(w).Encode(trans)
	if err != nil {
		return
	}
}

func GetTransByHash(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("content-type", "application/json")
	query := r.URL.Query()
	hash := query.Get("hash")
	transaction := QueryTranByHash(hash)
	if transaction.ID != 0 {
		_ = json.NewEncoder(w).Encode(transaction)
	} else {
		_ = json.NewEncoder(w).Encode(FormatResponse(500, "No such transaction"))
	}
}

func GetBalance(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("content-type", "application/json")
	vars := mux.Vars(r)
	key := vars["address"]
	var outBalance int
	var inBalance int
	for _, block := range QueryBlockChain() {
		for _, tran := range block.Transactions {
			if tran.From == key {
				outBalance += tran.Value

			}
			if tran.To == key {
				inBalance += tran.Value
			}
		}

	}
	var balance Balance
	balance.Address = key
	balance.Balance = inBalance - outBalance
	err := json.NewEncoder(w).Encode(balance)
	if err != nil {
		return
	}
}
