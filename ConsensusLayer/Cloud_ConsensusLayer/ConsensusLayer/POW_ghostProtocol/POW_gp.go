package main

import (
	"TinyChain/ConsensusLayer/General"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"TinyChain/Network/gossip"

	"TinyChain/Network/message"

	"github.com/cbergoon/merkletree"

	"sync"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/gorilla/mux"
)

const defaultPort = ":61613"

//the initial difficulty will be 0(1 actually, since will be increased when generate the fisrt block)
var difficulty = 0

//MaxDifficulty is the maximum difficulty you want to set
var MaxDifficulty int

var transIndex int

// Addresses : user addresses, the genesis node will grant them tokens
var Addresses []string

var address = ""
var next = ""

var g *gossip.Gossip
var CONN_HOST = "localhost"
var CONN_PORT = ""
var CONN_TYPE = "tcp"

var messageType = message.MessageType{"Bootstrap", "LatestBlockChain", "CurrentWinner", "NewTransaction", "NewProposedBlock", "NewJoinNode", "TEST"}

// Block represents each 'item' in the blockchain
type Block struct {
	Index      int
	Timestamp  string
	Hash       string
	PrevHash   string
	MerkleRoot []byte
	Difficulty int
	Nonce      string
	Signature  string
	Body       BlockBody
}
type State struct {
}

//BlockBody represents the formal expansion of the transactions
type BlockBody struct {
	Transactions []Transaction
}

// Response is the http response we send back to the client
type Response struct {
	Code    int    `json:"code"`
	Comment string `json:"comment"`
}

//Transaction : decoded transaction data converted from WalletMsg.Data
type Transaction struct {
	Date      string `json:"date"`
	From      string `json:"from"`
	ID        string `json:"id"`
	Signature string `json:"signature"`
	To        string `json:"to"`
	Value     int    `json:"value"`
}

func (t Transaction) CalculateHash() ([]byte, error) {
	h := sha256.New()
	if _, err := h.Write([]byte(t.Signature)); err != nil {
		return nil, err
	}
	return h.Sum(nil), nil
}

//Equals tests for equality of two Contents
func (t Transaction) Equals(other merkletree.Content) (bool, error) {
	return t.Signature == other.(Transaction).Signature, nil
}

// Balance struct is a struct which stores the balance of the user address
type Balance struct {
	Address string `json:"address"`
	Balance int    `json:"balance"`
}

// Config struct used to read json config
type Config struct {
	Strategy          string    `json:"Strategy"`
	RemoteStrategy    string    `json:"RemoteStrategy"`
	PublicAdd         string    `json:"PublicAdd"`
	PrivateAdd        string    `json:"PrivateAdd"`
	Port              string    `json:"Port"`
	ServerPort        string    `json:"ServerPort"`
	Addresses         []Address `json:"addresses"`
	TransactionIndex  int       `json:"transIndex"`
	MaximumDifficulty int       `json:"maximumDifficulty"`
}

// Address struct used to read json config
type Address struct {
	Address string `json:"address"`
}

type Detail struct {
	BlockChainHeight  int `json:"blockChainHeight"`
	TransactionNumber int `json:"TransactionNumber"`
}

//PublicAdd public cloud address
var PublicAdd string

// PrivateAdd private cloud address
var PrivateAdd string

// Port for gp
var Port string

var ServerPort string

//Transactions : a slice where we store the Transaction
var transactions []Transaction

//declinedTrans : a slice where we store the  declined Transaction
var declinedTrans []Transaction

// Blockchain is a series of validated Blocks
var Blockchain []Block
var tempBlocks []Block

var poolSize = 0

// announcements broadcasts winning validator to all nodes
var announcements = make(chan string)

// tranChan is a channel middleware for transaction
var tranChan = make(chan Transaction)

var mutex = &sync.Mutex{}

//all output will be made through this channel to clean up
var stdout = make(chan interface{})

// BootNodeAddress based on the strategy to settle the bootnode address
var BootNodeAddress string

func main() {
	readConfig()
	t := strconv.FormatInt(time.Now().Unix(), 10)
	genesisBlock := Block{}
	genesisBlockBody := BlockBody{}
	transaction1 := Transaction{t, "genesisTransaction", "0", General.CalculateHash(t + "genesisTransaction" + "0" + "Hao" + string(20)), Addresses[0], 500}

	//declare new transactions list and add the genesis trans in
	var genesisTransactions []Transaction
	genesisTransactions = append(genesisTransactions, transaction1)
	//genesisTransactions = append(genesisTransactions, transaction3)
	//convert into content list so that we can use the merkle tree package
	var genesisTransactionsContent = copyToContent(genesisTransactions)
	tr, err := merkletree.NewTree(genesisTransactionsContent)
	if err != nil {
		log.Fatal(err)
	}

	//blockbody
	genesisBlockBody = BlockBody{genesisTransactions}

	//calculate the nonce based on differentculty 0
	hex := fmt.Sprintf("%x", 0)
	genesisBlock = Block{
		0, t, calculateBlockHash(genesisBlock), "", tr.MerkleRoot(), 1, hex, "genesis", genesisBlockBody}

	spew.Dump(genesisBlock)
	Blockchain = append(Blockchain, genesisBlock)

	address = PublicAdd
	println("I am: " + address)

	CONN_PORT = Port
	g = gossip.NewGossip(PublicAdd, PrivateAdd, CONN_PORT, 3)
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

	go readInput(userInput)

	msg := message.JSONMessage{messageType.Bootstrap, PublicAdd + ":" + Port, time.Now().String(), address, 0}
	//send address to the bootnode
	g.SendDirect(msg, BootNodeAddress)

	go func() {
		for {
			select {
			case strs := <-userInput:
				if strings.HasPrefix("boot", strs) {
					fmt.Println("Sending message to BootNode")
					msg := message.JSONMessage{messageType.Bootstrap, PublicAdd + ":" + Port, time.Now().String(), address, 0}
					//send address to the bootnode
					//g.SendDirect(msg, "13.211.132.135:3333")
					g.SendDirect(msg, BootNodeAddress)
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
				fmt.Println("================================this is: " + address + "=======================================")
			}
		}
	}()

	// Init router
	r := mux.NewRouter()

	// Route handles & endpoints
	//get the transaction pool
	state := State{}
	r.HandleFunc("/getBlockchainStatus", state.getBlockchainStatus).Methods("GET")
	r.HandleFunc("/getBlockchain", state.getBlockchain).Methods("GET")
	r.HandleFunc("/getPartBlockchain", state.getPartBlockchain).Methods("GET")
	r.HandleFunc("/getTransaction/{address}", state.getUserTransactions).Methods("GET")
	r.HandleFunc("/sendTransaction", state.createTransaction).Methods("POST")
	r.HandleFunc("/getLastestBlock", state.getLatestBlock).Methods("GET")
	r.HandleFunc("/getBalance/{address}", state.getBalance).Methods("GET")
	r.HandleFunc("/getAllTransactions", state.getAllTransactions).Methods("GET")
	//get all the past transactions
	r.HandleFunc("/getDeclinedTransactions", getDeclinedTransactions).Methods("GET")

	//see if such transaction is in the specified block
	r.HandleFunc("/verifyTransaction/{blockIndex}", verifyTransaction).Methods("GET")

	// Start server
	log.Fatal(http.ListenAndServe(":"+ServerPort, r))

}

func readInput(input chan<- string) {
	for {
		var u string
		_, err := fmt.Scanf("%s\n", &u)
		if err != nil {
			panic(err)
		}
		input <- u
	}
}

func readConfig() {
	jsonFile, err := os.Open("PBFT.json")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Successfully Opened users.json")
	byteValue, _ := ioutil.ReadAll(jsonFile)

	var config Config
	err = json.Unmarshal(byteValue, &config)
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
	} else {
		BootNodeAddress = config.Strategy
	}
	PublicAdd = config.PublicAdd
	PrivateAdd = config.PrivateAdd
	Port = config.Port
	ServerPort = config.ServerPort
}

func getDeclinedTransactions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(declinedTrans)
}

func verifyTransaction(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	key := vars["blockIndex"]
	var transaction Transaction
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
	json.NewEncoder(w).Encode(msg)

}

func (state *State) getAllTransactions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("content-type", "application/json")
	var trans []Transaction
	for _, block := range Blockchain {
		trans = append(trans, block.Body.Transactions...)
	}
	json.NewEncoder(w).Encode(trans)
}

func (state *State) getUserTransactions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("content-type", "application/json")
	vars := mux.Vars(r)
	key := vars["address"]
	var trans []Transaction
	for _, block := range Blockchain {
		for _, tran := range block.Body.Transactions {
			if tran.From == key || tran.To == key {
				trans = append(trans, tran)
			}
		}
	}
	json.NewEncoder(w).Encode(trans)
}

func (state *State) getLatestBlock(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("content-type", "application/json")
	json.NewEncoder(w).Encode(Blockchain[len(Blockchain)-1])
}

func (state *State) getBlockchain(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("content-type", "application/json")
	json.NewEncoder(w).Encode(Blockchain)
}

func (state *State) getPartBlockchain(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("content-type", "application/json")
	query := r.URL.Query()
	pageIndex, _ := strconv.Atoi(query.Get("Index"))
	pageSize, _ := strconv.Atoi(query.Get("Size"))
	if pageIndex < 1 || (pageIndex-1)*pageSize >= len(Blockchain) {
		json.NewEncoder(w).Encode(General.FormatResponse(500, "Something goes wrong please try again"))
		return
	}
	var BlockchainPart []Block
	for i := (pageIndex - 1) * pageSize; i < pageIndex*pageSize; i++ {
		if i > len(Blockchain)-1 {
			break
		}
		BlockchainPart = append(BlockchainPart, Blockchain[i])
	}
	json.NewEncoder(w).Encode(BlockchainPart)
}

func (state *State) getBlockchainStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("content-type", "application/json")
	var trans []Transaction
	for _, block := range Blockchain {
		trans = append(trans, block.Body.Transactions...)
	}
	detail := Detail{len(Blockchain), len(trans)}
	json.NewEncoder(w).Encode(detail)
}

func (state *State) getBalance(w http.ResponseWriter, r *http.Request) {
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
	json.NewEncoder(w).Encode(balance)
}

func (state *State) createTransaction(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("content-type", "application/json")
	var transaction Transaction
	_ = json.NewDecoder(r.Body).Decode(&transaction)
	_, found := FindTrans(transactions, transaction)
	if found {
		json.NewEncoder(w).Encode(General.FormatResponse(500, "Something goes wrong please try again"))
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

	json.NewEncoder(w).Encode(General.FormatResponse(200, "transaction created successfully"))
}

//copy the transaction type to merkletree.Content type
func copyToContent(trans []Transaction) []merkletree.Content {
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
	// simulate receiving broadcast
	//fmt.Println(messageType)
	fmt.Println("send:" + messageType)
	m := message.JSONMessage{Type: messageType, Body: msg, Time: time.Now().String(), Signature: address}

	g.Broadcast(m)
}

func pickWinner() {
	// give 30 seconds each time to let the validators to propose the new block
	time.Sleep(10 * time.Second)
	//fmt.Println(transactions)
	if len(Blockchain) > 1 {
		updateTransaction(Blockchain[len(Blockchain)-1])
	} else if len(Blockchain) == 1 {
		updateTransaction(Blockchain[0])
	}

	var trans []Transaction

	//we only accept the  earliest transaction for each user at one block, so that to make sure the user will not double spend

	for _, tran := range transactions {
		if !checkSameIssuer(tran, trans) {
			trans = append(trans, tran)
		}
	}

	//perfrom the check on whether the user has enough balance to make the transaction
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

		//blockbody
		body = BlockBody{trans}

		//generate the block
		newBlock, err := generateBlock(oldLastIndex, body, address, tr.MerkleRoot())

		newBlockJSON, err := json.Marshal(newBlock)
		if err != nil {
			log.Fatal(err)
		}

		//check if the trans has already been verified, this may happen because other nodes has completed the verification
		//before this one does. if not we append the block to the blockchain. otherwise we discard this process
		for _, tran := range newBlock.Body.Transactions {
			if FindTransFromBlockchain(tran) {
				//if find any of the transaction is already verified, discard what we have
				return
			}
		}

		//other wise append the block
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

			//broadcast the lastest version blockchain
			broadcastBlockchain(string(output), messageType.LatestBlockChain)
		*/

	}

}

// generateBlock creates a new block using previous block's hash
func generateBlock(oldBlock Block, body BlockBody, address string, merkleRoot []byte) (Block, error) {

	var newBlock Block

	t := time.Now()

	newBlock.Index = oldBlock.Index + 1
	newBlock.Timestamp = t.String()
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

func checkSameIssuer(tran Transaction, trans []Transaction) bool {
	for _, t := range trans {
		if t.From == tran.From {
			return true
		}
	}
	return false
}

func checkBalance(trans []Transaction) []Transaction {
	var counter1 []int
	var counter2 []int
	for i, tran := range trans {
		if getMyBalance(tran.From) >= tran.Value {
			counter1 = append(counter1, i)
		} else {
			counter2 = append(counter2, i)
		}
	}
	var newTrans []Transaction
	for _, index := range counter1 {
		newTrans = append(newTrans, trans[index])
	}
	for _, index := range counter2 {
		declinedTrans = append(declinedTrans, trans[index])
	}
	removeTransFromPool(declinedTrans)

	return newTrans
}

func removeTransFromPool(trans []Transaction) {
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

// isBlockValid makes sure block is valid by checking index
// and comparing the hash of the previous block
func isBlockValid(newBlock, oldBlock Block) bool {
	if oldBlock.Index+1 != newBlock.Index {
		return false
	}

	if oldBlock.Hash != newBlock.PrevHash {
		return false
	}

	if calculateBlockHash(newBlock) != newBlock.Hash {
		return false
	}

	return true
}

//calculateBlockHash returns the hash of all block information
func calculateBlockHash(block Block) string {
	record := string(block.Index) + block.Timestamp + block.PrevHash + string(block.Difficulty) + block.Nonce + block.Signature + string(block.MerkleRoot) + calculateBlockBodyHash(block.Body)
	return General.CalculateHash(record)
}

//calculateBlockBodyHash returns the hash of all blockBody information
func calculateBlockBodyHash(blockBody BlockBody) string {
	record := ""
	for _, tran := range blockBody.Transactions {
		record = record + tran.Signature
	}
	return General.CalculateHash(record)
}

func isHashValid(hash string, difficulty int) bool {
	prefix := strings.Repeat("0", difficulty)
	return strings.HasPrefix(hash, prefix)
}

// Find takes a slice and looks for an element in it. If found it will
// return it's key, otherwise it will return -1 and a bool of false.
func Find(slice []string, val string) (int, bool) {
	for i, item := range slice {
		if item == val {
			return i, true
		}
	}
	return -1, false
}

// FindTrans takes a slice and looks for an element in it. If found it will
// return it's key, otherwise it will return -1 and a bool of false.
func FindTrans(slice []Transaction, val Transaction) (int, bool) {
	for i, item := range slice {
		if item.Signature == val.Signature {
			return i, true
		}
	}
	return -1, false
}

// FindTransFromBlockchain takes a transaction and looks for a same transaction in the current blockchain. If found it will
// return true, otherwise it will return false
func FindTransFromBlockchain(val Transaction) bool {
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
	var newTrans []Transaction
	for i := range transactions {
		if !FindIndex(counter, i) {
			newTrans = append(newTrans, transactions[i])
		}
	}

	transactions = newTrans

}

//handle msg
func msgHandler(msg message.JSONMessage) {

	//fmt.Println("Receive Message Type: " + msg.Type)
	//receive from the broadcast
	switch msg.Type {

	case messageType.NewTransaction:
		var transaction Transaction
		err := json.Unmarshal([]byte(msg.Body), &transaction)
		if err != nil {
			log.Fatal(err)
		}
		//fmt.Println(transaction)
		found := FindTransFromBlockchain(transaction)
		if !found {
			_, found2 := FindTrans(transactions, transaction)
			if !found2 {
				transactions = append(transactions, transaction)
				stdout <- "New transaction received and add"
				stdout <- transaction

			}
		}

		//fmt.Println(transactions)

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
				msg := message.JSONMessage{"NewJoinNode", g.PubAddr, time.Now().String(), "localhost:" + address, 0}
				g.SendDirect(msg, peer)
			}
		}

	}

}
