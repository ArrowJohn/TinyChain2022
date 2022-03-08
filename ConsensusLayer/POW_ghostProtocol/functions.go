package main

import (
	"TinyChain/ConsensusLayer/General"
	"TinyChain/Network/message"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/cbergoon/merkletree"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

func (t POWTransaction) CalculateHash() ([]byte, error) {
	h := sha256.New()
	if _, err := h.Write([]byte(t.Signature)); err != nil {
		return nil, err
	}
	return h.Sum(nil), nil
}

//Equals tests for equality of two Contents
func (t POWTransaction) Equals(other merkletree.Content) (bool, error) {
	return t.Signature == other.(POWTransaction).Signature, nil
}

func getDeclinedTransactions(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(declinedTrans)
	if err != nil {
		return
	}
}

func verifyTransaction(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	key := vars["blockIndex"]
	var transaction POWTransaction
	_ = json.NewDecoder(r.Body).Decode(&transaction)
	i, err := strconv.Atoi(key)
	if err != nil {
		log.Fatal()
	}
	b := Blockchain[i]
	trans := copyToContent(b.Body.Transactions)
	t, err := merkletree.NewTree(trans)
	if err != nil {
		log.Fatal(err)
	}
	vc, err := (t).VerifyContent(transaction)
	if err != nil {
		log.Fatal(err)
	}
	var msg string
	if vc {
		msg = "this transaction has been verified in the block " + key
	} else {
		msg = "this transaction is not verified in the block " + key
	}
	err = json.NewEncoder(w).Encode(msg)
	if err != nil {
		return
	}

}

func getAllTransFromChain() []POWTransaction {
	var trans []POWTransaction
	for _, block := range Blockchain {
		trans = append(trans, block.Body.Transactions...)
	}
	return trans
}

func (pow *POWGP) GetAllTransactions(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("content-type", "application/json")
	err := json.NewEncoder(w).Encode(getAllTransFromChain())
	if err != nil {
		return
	}
}

func (pow *POWGP) GetUserTransactions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("content-type", "application/json")
	vars := mux.Vars(r)
	key := vars["address"]
	var trans []POWTransaction
	for _, block := range Blockchain {
		for _, tran := range block.Body.Transactions {
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

func (pow *POWGP) GetLatestBlock(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("content-type", "application/json")
	err := json.NewEncoder(w).Encode(Blockchain[len(Blockchain)-1])
	if err != nil {
		return
	}
}

func (pow *POWGP) GetBlockchain(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("content-type", "application/json")
	err := json.NewEncoder(w).Encode(Blockchain)
	if err != nil {
		return
	}
}

func (pow *POWGP) GetPartBlockchain(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("content-type", "application/json")
	query := r.URL.Query()
	pageIndex, _ := strconv.Atoi(query.Get("Index"))
	pageSize, _ := strconv.Atoi(query.Get("Size"))
	if pageIndex < 1 || (pageIndex-1)*pageSize >= len(Blockchain) {
		err := json.NewEncoder(w).Encode(General.FormatResponse(500, "Something goes wrong please try again"))
		if err != nil {
			return
		}
		return
	}
	var BlockchainPart []Block
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

func (pow *POWGP) GetBlockByIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("content-type", "application/json")
	query := r.URL.Query()
	index, _ := strconv.Atoi(query.Get("index"))
	for _, block := range Blockchain {
		if block.Index == index {
			err := json.NewEncoder(w).Encode(block)
			if err != nil {
				return
			}
			return
		}
	}
	err := json.NewEncoder(w).Encode(General.FormatResponse(500, "No such block"))
	if err != nil {
		return
	}
}

func (pow *POWGP) GetBlockchainStatus(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("content-type", "application/json")
	detail := Detail{len(Blockchain), len(getAllTransFromChain())}
	err := json.NewEncoder(w).Encode(detail)
	if err != nil {
		return
	}
}

func (pow *POWGP) GetBalance(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("content-type", "application/json")
	vars := mux.Vars(r)
	key := vars["address"]
	var outBalance int
	var inBalance int
	for _, block := range Blockchain {
		for _, tran := range block.Body.Transactions {
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

func (pow *POWGP) GetTransByHash(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("content-type", "application/json")
	query := r.URL.Query()
	hash := query.Get("hash")
	var trans []POWTransaction
	for _, block := range Blockchain {
		trans = append(trans, block.Body.Transactions...)
	}
	transaction := General.FindTransByHash(unifyTransactions(trans), hash)
	if transaction.Date == "" {
		err := json.NewEncoder(w).Encode(General.FormatResponse(500, "No such transaction"))
		if err != nil {
			return
		}
		return
	}
	err := json.NewEncoder(w).Encode(transaction)
	if err != nil {
		return
	}

}

func (pow *POWGP) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("content-type", "application/json")
	var transaction POWTransaction
	_ = json.NewDecoder(r.Body).Decode(&transaction)
	_, found := General.FindTrans(unifyTransactions(transactions), unifyTransaction(transaction))
	if found {
		err := json.NewEncoder(w).Encode(General.FormatResponse(500, "Something goes wrong please try again"))
		if err != nil {
			return
		}
		return
	}
	//append the transaction to the channel first then retrieve it later
	//tranChan <- transaction
	transactions = append(transactions, transaction)
	stdout <- "Receive user sent Transactions"
	stdout <- transaction
	output, err := json.Marshal(transaction)
	if err != nil {
		log.Fatal(err)
	}

	//broadcast this transaction to the P2P network
	broadcastBlockchain(string(output), messageType.NewTransaction)

	err = json.NewEncoder(w).Encode(General.FormatResponse(200, "transaction created successfully"))
	if err != nil {
		return
	}
}

func (pow *POWGP) GetCurrentTrans(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("content-type", "application/json")
	query := r.URL.Query()
	user := query.Get("address")
	days, _ := strconv.Atoi(query.Get("days"))
	var trans []POWTransaction
	for _, block := range Blockchain {
		for _, tran := range block.Body.Transactions {
			if (tran.From == user || tran.To == user) && General.JudgeInDays(tran.Date, days) {
				trans = append(trans, tran)
			}
		}
	}
	err := json.NewEncoder(w).Encode(trans)
	if err != nil {
		return
	}
}

//copy the transaction type to merkletree.Content type
func copyToContent(trans []POWTransaction) []merkletree.Content {
	var list []merkletree.Content
	for _, tran := range trans {
		list = append(list, tran)
	}
	return list
}

// Get balance
func getMyBalance(address string) int {
	var outBalance int
	var inBalance int
	for _, block := range Blockchain {
		for _, tran := range block.Body.Transactions {
			if tran.From == address {
				outBalance += tran.Value

			}
			if tran.To == address {
				inBalance += tran.Value
			}
		}

	}
	var balance Balance
	balance.Address = address
	balance.Balance = inBalance - outBalance
	return balance.Balance

}

func broadcastBlockchain(msg string, messageType string) {
	fmt.Println("send:" + messageType)
	m := message.JSONMessage{Type: messageType, Body: msg, Time: time.Now().String(), Signature: address}

	g.Broadcast(m)
}

func pickWinner() {
	// give 30 seconds each time to let the validators to propose the new block
	time.Sleep(10 * time.Second)
	if len(Blockchain) > 1 {
		updateTransaction(Blockchain[len(Blockchain)-1])
	} else if len(Blockchain) == 1 {
		updateTransaction(Blockchain[0])
	}

	var trans []POWTransaction

	//we only accept the  earliest transaction for each user at one block, so that to make sure the user will not double spend

	for _, tran := range transactions {
		if !checkSameIssuer(tran, trans) {
			trans = append(trans, tran)
		}
	}

	//perform the check on whether the user has enough balance to make the transaction
	trans = checkBalance(trans)
	stdout <- "Current Transactions Pool"
	stdout <- transactions
	stdout <- "Accepted Transactions pool:"
	stdout <- trans
	stdout <- "Declined Transactions pool:"
	stdout <- declinedTrans

	if len(trans) >= transIndex {
		stdout <- "start generating Block: "
		//generate the block
		//fmt.Println("generating block")
		mutex.Lock()
		oldLastIndex := Blockchain[len(Blockchain)-1]
		mutex.Unlock()

		body := BlockBody{}

		//convert into content list so that we can use the merkle tree package
		var TransContent = copyToContent(trans)
		tr, err := merkletree.NewTree(TransContent)
		if err != nil {
			log.Fatal(err)
		}
		id := len(getAllTransFromChain())
		for i, item := range trans {
			if item.Hash == "" {
				trans[i].Hash = General.CalculateTranHash(unifyTransaction(item))
			}
			if item.ID == "" {
				trans[i].ID = strconv.Itoa(id)
				id++
			}

		}
		//BlockBody
		body = BlockBody{trans}

		//generate the block
		newBlock, err := generateBlock(oldLastIndex, body, address, tr.MerkleRoot())

		newBlockJSON, err := json.Marshal(newBlock)
		if err != nil {
			log.Fatal(err)
		}

		//check if the trans has already been verified, this may happen because other nodes has completed the verification
		//before this one does. if not we append the block to the blockchain. otherwise, we discard this process
		for _, tran := range newBlock.Body.Transactions {
			if FindTransFromBlockchain(tran) {
				//if you find any of the transaction is already verified, discard what we have
				return
			}
		}

		//otherwise, append the block
		broadcastBlockchain(string(newBlockJSON), messageType.NewProposedBlock)
		stdout <- "New Proposed Block send out"
		stdout <- newBlock
		Blockchain = append(Blockchain, newBlock)
		updateTransaction(newBlock)
		/*
			output, err := json.Marshal(Blockchain)
			if err != nil {
				log.Fatal(err)
			}
			//remove the transactions that has been verified from the transactions pool

			//broadcast the latest version blockchain
			broadcastBlockchain(string(output), messageType.LatestBlockChain)
		*/

	}

}

// generateBlock creates a new block using previous block's hash
func generateBlock(oldBlock Block, body BlockBody, address string, merkleRoot []byte) (Block, error) {
	var newBlock Block

	newBlock.Index = oldBlock.Index + 1
	newBlock.Timestamp = General.CurrentTimestamp()
	newBlock.Body = body
	newBlock.PrevHash = oldBlock.Hash
	if difficulty < MaxDifficulty {
		difficulty++
		newBlock.Difficulty = difficulty
	} else if difficulty == MaxDifficulty {
		newBlock.Difficulty = difficulty
		difficulty++
	} else {
		difficulty = 0
		newBlock.Difficulty = difficulty + 1
	}

	newBlock.Signature = address
	newBlock.MerkleRoot = merkleRoot

	for i := 0; ; i++ {
		hex := fmt.Sprintf("%x", i)
		newBlock.Nonce = hex
		//check if the hash(with the appended nonce) has the amount of zero that equals to the difficulty
		if !isHashValid(calculateBlockHash(newBlock), newBlock.Difficulty) {
			stdout <- "Generating.........."
			stdout <- calculateBlockHash(newBlock)
			time.Sleep(time.Second)
			continue
		} else {
			stdout <- "Work Done!"
			stdout <- calculateBlockHash(newBlock)
			newBlock.Hash = calculateBlockHash(newBlock)
			break
		}

	}

	return newBlock, nil
}

func isHashValid(hash string, difficulty int) bool {
	prefix := strings.Repeat("0", difficulty)
	return strings.HasPrefix(hash, prefix)
}

func checkSameIssuer(tran POWTransaction, trans []POWTransaction) bool {
	for _, t := range trans {
		if t.From == tran.From {
			return true
		}
	}
	return false
}

func checkBalance(trans []POWTransaction) []POWTransaction {
	var counter1 []int
	var counter2 []int
	for i, tran := range trans {
		if getMyBalance(tran.From) >= tran.Value {
			counter1 = append(counter1, i)
		} else {
			counter2 = append(counter2, i)
		}
	}
	var newTrans []POWTransaction
	for _, index := range counter1 {
		newTrans = append(newTrans, trans[index])
	}
	for _, index := range counter2 {
		declinedTrans = append(declinedTrans, trans[index])
	}
	General.RemoveTransFromPool(unifyTransactions(declinedTrans), unifyTransactions(declinedTrans))
	return newTrans
}

//calculateBlockHash returns the hash of all block information
func calculateBlockHash(block Block) string {
	record := string(rune(block.Index)) + block.Timestamp + block.PrevHash +
		string(rune(block.Difficulty)) + block.Nonce + block.Signature +
		string(block.MerkleRoot) + General.CalculateTransHash(unifyTransactions(block.Body.Transactions))
	return General.CalculateHash(record)
}

// FindTransFromBlockchain takes a transaction and looks for a same transaction in the current blockchain. If found it will
// return true, otherwise it will return false
func FindTransFromBlockchain(val POWTransaction) bool {
	//Verify a specific content in the tree
	//we only need to verify this by checking the merkle root instead of looping through the whole transactions list
	for _, item := range Blockchain {
		trans := copyToContent(item.Body.Transactions)
		t, err := merkletree.NewTree(trans)
		if err != nil {
			log.Fatal(err)
		}
		vc, err := (t).VerifyContent(val)
		if err != nil {
			log.Fatal(err)
		}
		if vc {
			return vc
		}
	}

	return false

}

// FindIndex  takes a slice and looks for an element in it. If found it will
// return it's key, otherwise it will return -1 and a bool of false.
func FindIndex(slice []int, val int) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

//update the transaction based on the transactions stored in the block
func updateTransaction(block Block) {
	var counter []int
	//takes O(n^3), not a good way
	for i, tran := range transactions {
		for _, storedTran := range block.Body.Transactions {
			if tran.Signature == storedTran.Signature {
				if !FindIndex(counter, i) {
					counter = append(counter, i)
				}
			}
		}
	}
	var newTrans []POWTransaction
	for i := range transactions {
		if !FindIndex(counter, i) {
			newTrans = append(newTrans, transactions[i])
		}
	}
	transactions = newTrans
}

func unifyTransactions(transactions []POWTransaction) []General.Transaction {
	var transactionsGeneral []General.Transaction
	for _, item := range transactions {
		transactionsGeneral = append(transactionsGeneral, unifyTransaction(item))
	}
	return transactionsGeneral
}

func unifyTransaction(transaction POWTransaction) General.Transaction {
	transactionGeneral := General.Transaction{}
	data, _ := json.Marshal(transaction)
	_ = json.Unmarshal(data, &transactionGeneral)
	return transactionGeneral
}

//handle msg
func msgHandler(msg message.JSONMessage) {

	//fmt.Println("Receive Message Type: " + msg.Type)
	//receive from the broadcast
	switch msg.Type {

	case messageType.NewTransaction:
		var transaction POWTransaction
		err := json.Unmarshal([]byte(msg.Body), &transaction)
		if err != nil {
			log.Fatal(err)
		}
		//fmt.Println(transaction)
		found := FindTransFromBlockchain(transaction)
		if !found {
			_, found2 := General.FindTrans(unifyTransactions(transactions), unifyTransaction(transaction))
			if !found2 {
				transactions = append(transactions, transaction)
				stdout <- "New transaction received and add"
				stdout <- transaction

			}
		}
		//broadcast this transaction to the P2P network
		//broadcastBlockchain(string(output), messageType.NewTransaction)

	case messageType.LatestBlockChain:
		//update the transaction pool

		//println("Latest Blockchain Received From:" + m.Signature)
		var nextBlockChain []Block
		err := json.Unmarshal([]byte(msg.Body), &nextBlockChain)
		if err != nil {
			log.Fatal(err)
		}
		//fmt.Println(nextBlockChain)
		if len(nextBlockChain) > len(Blockchain) {
			stdout <- "Replaced by a longer chain"
			Blockchain = nextBlockChain
			//takes O(N^3), not a good practise, but currently stick to this
			for _, b := range Blockchain {
				updateTransaction(b)
			}

		} else if len(nextBlockChain) == len(Blockchain) {
			//fmt.Println("length same")
			//do nothing if lengths are the same
		} else {
			//current nodes has the longest chain, so it broadcast its version to the globe
			// marshal the blockchain into JSON format

		}

	case messageType.NewProposedBlock:
		var block Block
		err := json.Unmarshal([]byte(msg.Body), &block)
		if err != nil {
			log.Fatal(err)
		}
		//verify the block.
		hash := calculateBlockHash(block)
		if isHashValid(hash, block.Difficulty) {
			if hash == block.Hash {
				//block verified add to the blockchain
				stdout <- "Block received, verified, and append to the Blockchain"
				stdout <- block
				Blockchain = append(Blockchain, block)
				updateTransaction(block)
			}
		}

	case messageType.NewJoinNode:
		stdout <- "[gossip] Received a NewJoinNode message(add):" + msg.Body
		err := g.AddPeer(msg.Body)
		if err == nil {
			g.NewMsg <- msg
		}

	case messageType.Bootstrap:
		stdout <- "[gossip] Bootstrap received"
		peerList := strings.Split(msg.Body, ",")

		for _, peer := range peerList {
			err := g.AddPeer(peer)
			if err == nil {
				msg := message.JSONMessage{Type: "NewJoinNode", Body: g.PubAddr, Time: time.Now().String(), Signature: "localhost:" + address}
				err := g.SendDirect(msg, peer)
				if err != nil {
					return
				}
			}
		}

	}

}

func readConfig() {
	jsonFile, err := os.Open("pow.json")
	if err != nil {
		fmt.Println(err)
		jsonFile, _ = os.Open("ConsensusLayer/POW_ghostProtocol/pow.json")
	}
	byteValue, _ := ioutil.ReadAll(jsonFile)

	var config Config
	err = json.Unmarshal(byteValue, &config)
	fmt.Println(config)
	print(config.Addresses)
	if err != nil {
		return
	}
	for i := 0; i < len(config.Addresses); i++ {
		Addresses = append(Addresses, config.Addresses[i].Address)
	}
	transIndex = config.TransactionIndex
	MaxDifficulty = config.MaximumDifficulty

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
