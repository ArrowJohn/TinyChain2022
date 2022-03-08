package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/cbergoon/merkletree"

	"sync"
	"time"

	"TinyChain/Network/message"
	"TinyChain/Network/pubsub"
	"github.com/davecgh/go-spew/spew"
	"github.com/gorilla/mux"
)

const defaultPort = ":61613"

var ps = pubsub.Connect("")

//var ps = pubsub.Connect("52.62.115.32:61613")

//the initial difficulty will be 0(1 actually, since will be increased when generate the fisrt block)
var difficulty = 0

//MaxDiffculty is the maximum difficulty you want to set
var MaxDiffculty int

var transIndex int

// Addresses : user addresses, the genesis node will grant them tokens
var Addresses []string

var address = ""
var next = ""

//all output will be made through this channel to clean up
var stdout = make(chan interface{})

var broadcast = flag.String("public", "/public/blockchain", "public")
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

//JSONMessage : define the JSON formate message type
type JSONMessage struct {
	Type      string
	Body      string
	Time      string
	Signature string
}

// Balance struct is a struct which stores the balance of the user address
type Balance struct {
	Address string `json:"address"`
	Balance int    `json:"balance"`
}

// Config struct used to read json config
type Config struct {
	Strategy         string    `json:"Strategy"`
	Addresses        []Address `json:"addresses"`
	TransactionIndex int       `json:"transIndex"`
	MaximumDiffculty int       `json:"maximumDiffculty"`
}

// Node struct used to read json config
type Node struct {
	Node string `json:"node"`
}

// Address struct used to read json config
type Address struct {
	Address string `json:"address"`
}

// NodesList that will be joining the P2P network
var NodesList []string

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

func main() {

	readConfig()
	// create genesis block
	t := "0000000"
	genesisBlock := Block{}
	transaction1 := Transaction{}
	transaction2 := Transaction{}
	//transaction3 := Transaction{}
	genesisBlockBody := BlockBody{}

	//give balance to the account in genesis block
	transaction1 = Transaction{t, "genesisTransaction", "0", calculateHash(t + "genesisTransaction" + "0" + "Hao" + string(20)), Addresses[0], 250}
	transaction2 = Transaction{t, "genesisTransaction", "0", calculateHash(t + "genesisTransaction" + "0" + "Hao" + string(20)), Addresses[1], 250}
	//transaction3 = Transaction{t, "genesisTransaction", "0", calculateHash(t + "genesisTransaction" + "0" + "Hao" + string(20)), Addresses[2], 20}

	//declare new transactions list and add the genesis trans in
	var genesisTransactions []Transaction
	genesisTransactions = append(genesisTransactions, transaction1)
	genesisTransactions = append(genesisTransactions, transaction2)
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
	genesisBlock = Block{0, t, calculateBlockHash(genesisBlock), "", tr.MerkleRoot(), 1, hex, "genesis", genesisBlockBody}

	spew.Dump(genesisBlock)
	Blockchain = append(Blockchain, genesisBlock)

	//parse command-line args
	//target := flag.String("l", "", "target peer to dial")
	port := flag.String("p", "", "node address")
	server := flag.String("s", "", "server port for restful api")
	flag.Parse()

	if *port == "" {
		log.Fatal("port is invalid")
	}

	address = *port
	println("I am: " + address)

	ps.Subscribe(*broadcast, msgHandler)

	//choose the validator
	go func() {
		for {
			pickWinner()
		}
	}()

	go func() {
		for {
			//periodally broadcast the current blockchain
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
	r.HandleFunc("/getTransactions", getTransactions).Methods("GET")
	r.HandleFunc("/getTransaction/{address}", getUserTransactions).Methods("GET")

	//open a separate goroutine to handle the transaction creation
	r.HandleFunc("/sendTransaction", createTransaction).Methods("POST")

	r.HandleFunc("/getBlockchain", getBlockchain).Methods("GET")
	r.HandleFunc("/getLastestBlock", getLatestBlock).Methods("GET")
	r.HandleFunc("/getBalance/{address}", getBalance).Methods("GET")

	//get all the past transactions
	r.HandleFunc("/getAllTransactions", getAllTransactions).Methods("GET")

	//get all the past transactions
	r.HandleFunc("/getDeclinedTransactions", getDeclinedTransactions).Methods("GET")

	//see if such transaction is in the specified block
	r.HandleFunc("/verifyTransaction/{blockIndex}", verifyTransaction).Methods("GET")

	// Start server
	log.Fatal(http.ListenAndServe(":"+*server, r))

}

func readConfig() {
	// Open our jsonFile
	jsonFile, err := os.Open("config.json")
	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Successfully Opened users.json")
	// defer the closing of our jsonFile so that we can parse it later on
	// read our opened jsonFile as a byte array.
	byteValue, _ := ioutil.ReadAll(jsonFile)

	// we initialize our Users array
	var config Config

	// we unmarshal our byteArray which contains our
	// jsonFile's content into 'nodes' and 'transIndex' which we defined above
	json.Unmarshal(byteValue, &config)
	for i := 0; i < len(config.Addresses); i++ {
		Addresses = append(Addresses, config.Addresses[i].Address)
	}
	transIndex = config.TransactionIndex
	MaxDiffculty = config.MaximumDiffculty
	fmt.Println(config)
	if config.Strategy == "local" {
		ps = pubsub.Connect("localhost:61613")
	} else {
		ps = pubsub.Connect(config.Strategy)
	}

}

// Get all transactions
func getTransactions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(transactions)
}

func getDeclinedTransactions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(declinedTrans)

}

// Get all the past transactions
func getAllTransactions(w http.ResponseWriter, r *http.Request) {
	var trans []Transaction
	for _, block := range Blockchain {
		trans = append(trans, block.Body.Transactions...)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(trans)
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

// Get all transactions
func getUserTransactions(w http.ResponseWriter, r *http.Request) {
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
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(trans)
}

// Get the latest blockchain
func getLatestBlock(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Blockchain[len(Blockchain)-1])
}

// Get all blockchain
func getBlockchain(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Blockchain)
}

//copy the transaction type to merkletree.Content type
func copyToContent(trans []Transaction) []merkletree.Content {
	var list []merkletree.Content
	for _, tran := range trans {
		list = append(list, tran)
	}
	return list
}

//Equals tests for equality of two Contents
func (t Transaction) Equals(other merkletree.Content) (bool, error) {
	return t.Signature == other.(Transaction).Signature, nil
}

// Get balance
func getBalance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["address"]
	var outBalance int
	var inBalance int
	//get the transaction
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
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(balance)
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

// Add new transaction
func createTransaction(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var transaction Transaction
	_ = json.NewDecoder(r.Body).Decode(&transaction)

	//fmt.Println("Transaction JSON\n")
	//fmt.Fprintf(os.Stdout, "%s", trans)
	//fmt.Println("\n")

	_, found := FindTrans(transactions, transaction)
	if found {
		json.NewEncoder(w).Encode(formatResponse(500, "Something goes wrong please try again"))
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(formatResponse(200, "transaction created successfully"))
}

// Add new transaction
func formatResponse(code int, comment string) Response {
	response := Response{}
	response = Response{code, comment}
	return response
}

func broadcastBlockchain(msg string, messageType string) {
	// simulate receiving broadcast
	//fmt.Println(messageType)
	//fmt.Println("send:" + messageType)
	m := JSONMessage{messageType, msg, time.Now().String(), address}

	ps.Publish(*broadcast, marshalToJSON(m))
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

		//generate the block(POW)
		newBlock, err := generateBlock(oldLastIndex, body, address, tr.MerkleRoot())

		newBlockJSON, err := json.Marshal(newBlock)
		if err != nil {
			log.Fatal(err)
		}

		//check if the trans has already been verified, this may happen because other nodes has completed the verification
		//before this one does. if not we append the block to the blockchain. otherwise we discard this process
		for _, tran := range newBlock.Body.Transactions {
			if FindTransFromBlockchain(tran) {

				return
			}

		}
		broadcastBlockchain(string(newBlockJSON), messageType.NewProposedBlock)
		stdout <- "New Proposed Block send out"
		stdout <- newBlock
		Blockchain = append(Blockchain, newBlock)
		/*
			output, err := json.Marshal(Blockchain)
			if err != nil {
				log.Fatal(err)
			}
			broadcastBlockchain(string(output), messageType.LatestBlockChain)
		*/
		//remove the transactions that has been verified from the transactions pool
		updateTransaction(newBlock)
		//broadcast the lastest version blockchain

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
	if difficulty < MaxDiffculty {
		difficulty++
		newBlock.Difficulty = difficulty
	} else if difficulty == MaxDiffculty {
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

func marshalToJSON(message JSONMessage) string {
	output, err := json.Marshal(message)
	if err != nil {
		log.Fatal(err)
	}
	return string(output)
}

func unMarshalToString(message []byte) JSONMessage {
	var m JSONMessage
	err := json.Unmarshal([]byte(message), &m)
	if err != nil {
		log.Fatal(err)
	}
	return m
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

// SHA256 hasing
// calculateHash is a simple SHA256 hashing function
func calculateHash(s string) string {
	h := sha256.New()
	h.Write([]byte(s))
	hashed := h.Sum(nil)
	return hex.EncodeToString(hashed)
}

//CalculateHash hashes the values of a TestContent
func (t Transaction) CalculateHash() ([]byte, error) {
	h := sha256.New()
	if _, err := h.Write([]byte(t.Signature)); err != nil {
		return nil, err
	}

	return h.Sum(nil), nil
}

//calculateBlockHash returns the hash of all block information
func calculateBlockHash(block Block) string {
	record := string(block.Index) + block.Timestamp + block.PrevHash + string(block.Difficulty) + block.Nonce + block.Signature + string(block.MerkleRoot) + calculateBlockBodyHash(block.Body)
	return calculateHash(record)
}

//calculateBlockBodyHash returns the hash of all blockBody information
func calculateBlockBodyHash(blockBody BlockBody) string {
	record := ""
	for _, tran := range blockBody.Transactions {
		record = record + tran.Signature
	}
	return calculateHash(record)
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
	/*
		for i, item := range Blockchain {
			if item.Transaction.Signature == val.Signature && item.Transaction.Date == val.Date && item.Transaction.From == val.From && item.Transaction.ID == val.ID && item.Transaction.Value == val.Value {
				return i, true
			}
		}
		return -1, false
	*/

	//Verify a specific content in in the tree
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

// Find takes a slice and looks for an element in it. If found it will
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
func msgHandler(msgObj interface{}, topicObj interface{}) {
	m := msgObj.([]byte)
	//topic := topicObj.(string)
	msg := unMarshalToString([]byte(m))
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
		output, err := json.Marshal(transaction)
		if err != nil {
			log.Fatal(err)
		}
		//broadcast this transaction to the P2P network
		broadcastBlockchain(string(output), messageType.NewTransaction)

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
			fmt.Println(Blockchain)

		} else if len(nextBlockChain) == len(Blockchain) {
			//fmt.Println("length same")
			//do nothing if lengths are the same
		} else {
			//current nodes has the longest chain, so it broadcast its version to the globe
			// marshal the blockchain into JSON format

		}
		output, err := json.Marshal(Blockchain)
		if err != nil {
			log.Fatal(err)
		}
		broadcastBlockchain(string(output), messageType.LatestBlockChain)

	case messageType.NewProposedBlock:
		var block Block
		err := json.Unmarshal([]byte(msg.Body), &block)
		if err != nil {
			log.Fatal(err)
		}
		//verify the block.
		prefix := strings.Repeat("0", block.Difficulty)
		hash := string(block.Index) + block.Timestamp + block.PrevHash + string(block.Difficulty) + block.Nonce + block.Signature + string(block.MerkleRoot) + calculateBlockBodyHash(block.Body)
		if strings.HasPrefix(prefix, calculateHash(hash)) {
			if calculateHash(hash) == block.Hash {
				stdout <- "Block received, verified, and append to the Blockchain"
				stdout <- block
				//block verified add to the blockchain
				Blockchain = append(Blockchain, block)
				updateTransaction(block)
			}
		}
		broadcastBlockchain(msg.Body, messageType.NewProposedBlock)

		/*
			case messageType.NewJoinNode:
				//println("add node :" + m.Body)
				// append the next node into the validator if not exist in the current pool

				output, err := json.Marshal(Blockchain)
				if err != nil {
					log.Fatal(err)
				}
				//fmt.Println(validators)
				broadcastBlockchain(string(output), messageType.LatestBlockChain)
		*/

	}

}
