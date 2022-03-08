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
	"sort"

	"TinyChain/Network/pubsub"

	"sync"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/gorilla/mux"
)

const defaultPort = ":61613"

var address = ""
var next = ""

//var ps = pubsub.Connect("localhost:61613")
//var ps = pubsub.Connect("52.62.115.32:61613")
var ps = pubsub.Connect("")
var broadcast = flag.String("public", "/public/blockchain", "public")
var stop = make(chan bool)

// Block represents each 'item' in the blockchain
type Block struct {
	Index         int         `json:"index"`
	Timestamp     string      `json:"timestamp"`
	Transaction   Transaction `json:"transaction"`
	Hash          string      `json:"hash"`
	PrevHash      string      `json:"prevHash"`
	Validator     string      `json:"validator"`
	NextValidator string      `json:"nextValidator"`
}

// Response is the response we send back to the rerquest user
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

// Config struct used to read json config
type Config struct {
	Strategy   string    `json:"Strategy"`
	PoolLimit  int       `json:"PoolLimit"`
	Addresses  []Address `json:"addresses"`
	ServerPort string    `json:"ServerPort"`
	PublicAdd  string    `json:"PublicAdd"`
}

var ServerPort string

var PublicAdd string

// NodesList that will be joining the P2P network
var NodesList []string

// Addresses : user addresses, the genesis node will grant them tokens
var Addresses []string

// Address struct used to read json config
type Address struct {
	Address string `json:"address"`
}

//MessageType : define the types of the messages we have
type MessageType struct {
	NewJoinNode      string
	LatestBlockChain string
	CurrentWinner    string
	NewTransaction   string
	NewProposedBlock string
}

//Validator : define the types of the Validator we have
type Validator struct {
	Validator string `json:"Validator"`
	Timestamp int    `json:"timestamp"`
}

// Balance struct is a struct which stores the balance of the user address
type Balance struct {
	Address string `json:"address"`
	Balance int    `json:"balance"`
}

//Transactions : a slice where we store the Transaction
var transactions []Transaction

var messageType = MessageType{"NewJoinNode", "LatestBlockChain", "CurrentWinner", "NewTransaction", "NewProposedBlock"}

// Blockchain is a series of validated Blocks
var Blockchain []Block
var tempBlocks []Block
var validators []Validator
var currentValidator = ""

var poolLimit = 3

// announcements broadcasts winning validator to all nodes
var announcements = make(chan string)

//declinedTrans : a slice where we store the  declined Transaction
var declinedTrans []Transaction

// tranChan is a channel middleware for transaction
var tranChan = make(chan Transaction)

var mutex = &sync.Mutex{}

var stdout = make(chan interface{})

func main() {

	readConfig()

	// create genesis block
	t := "1606695587"
	genesisBlock := Block{}
	tra := Transaction{}

	//give balance to the account in genesis block
	tra = Transaction{t, "genesisTransaction", "0", calculateHash(t + "genesisTransaction" + "0" + "Hao" + string(20)), Addresses[0], 500}

	genesisBlock = Block{0, t, tra, "Genesis transaction: 06/10/2020", calculateBlockHash(genesisBlock), "", "8891"}

	spew.Dump(genesisBlock)
	if len(Blockchain) == 0 {
		Blockchain = append(Blockchain, genesisBlock)
	}
	//parse command-line args
	/*
		port := flag.String("p", "", "node address")
		server := flag.String("s", "", "server port for restful api")
		flag.Parse()

		if *port == "" || *server == "" {
			log.Fatal("port/vcalidator is invalid")
		}
	*/

	address = PublicAdd

	//subscribe the public channel
	ps.Subscribe(*broadcast, msgHandler)

	thisValidator := Validator{address, int(time.Now().Unix())}
	validators = append(validators, thisValidator)
	output, err := json.Marshal(thisValidator)
	if err != nil {
		log.Fatal(err)
	}
	broadcastBlockchain(string(output), messageType.NewJoinNode)

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

	//choose the validator
	go func() {
		for {
			pickWinner()
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

	//see if such transaction is in the specified block
	r.HandleFunc("/verifyTransaction/", verifyTransaction).Methods("GET")

	//get all the past transactions
	r.HandleFunc("/getDeclinedTransactions", getDeclinedTransactions).Methods("GET")
	// Start server
	log.Fatal(http.ListenAndServe(":"+ServerPort, r))

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

	poolLimit = config.PoolLimit
	fmt.Println(poolLimit)

	fmt.Println(config)
	if config.Strategy == "local" {
		ps = pubsub.Connect("localhost:61613")
	} else {
		ps = pubsub.Connect(config.Strategy + defaultPort)
	}

	ServerPort = config.ServerPort
	PublicAdd = config.PublicAdd

}

// Get all transactions
func getTransactions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(transactions)
}

// Get all the past transactions
func getAllTransactions(w http.ResponseWriter, r *http.Request) {
	var trans []Transaction
	for _, block := range Blockchain {
		trans = append(trans, block.Transaction)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(trans)
}

// Get all transactions
func getUserTransactions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["address"]
	var trans []Transaction
	for _, block := range Blockchain {
		if block.Transaction.From == key || block.Transaction.To == key {
			trans = append(trans, block.Transaction)
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

func getDeclinedTransactions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(declinedTrans)

}

// Get all blockchain
func getBlockchain(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(Blockchain)
}

func verifyTransaction(w http.ResponseWriter, r *http.Request) {
	var msg string
	w.Header().Set("Content-Type", "application/json")
	var transaction Transaction
	_ = json.NewDecoder(r.Body).Decode(&transaction)
	for _, b := range Blockchain {
		if b.Transaction.Signature == transaction.Signature {
			msg = "this transaction has been verified in the block "
			break
		}
		msg = "this transaction is not verified in the block "
	}

	json.NewEncoder(w).Encode(msg)

}

// Get balance
func getBalance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["address"]
	var outBalance int
	var inBalance int
	for _, block := range Blockchain {
		if block.Transaction.From == key {
			outBalance += block.Transaction.Value

		}
		if block.Transaction.To == key {
			inBalance += block.Transaction.Value
		}
	}
	var balance Balance
	balance.Address = key
	balance.Balance = inBalance - outBalance
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(balance)
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
		json.NewEncoder(w).Encode(formatResponse(500, "already exists in the transactions"))
		return
	}
	//append the transaction to the channel first then retrieve it later
	//tranChan <- transaction
	_, found2 := FindTrans(declinedTrans, transaction)
	if found2 {
		json.NewEncoder(w).Encode(formatResponse(500, "same transaction has been declined"))
		return
	}

	_, found3 := FindTransFromBlockchain(transaction)
	if found3 {
		json.NewEncoder(w).Encode(formatResponse(500, "Already exists in the blockchain"))
		return
	}
	transactions = append(transactions, transaction)
	//fmt.Println("send new trans")
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

func broadcastBlockchain(message string, messageType string) {
	// simulate receiving broadcast
	m := JSONMessage{messageType, message, time.Now().String(), address}
	//fmt.Println("Broadcast to: " + *broadcast)
	go ps.Publish(*broadcast, marshalToJSON(m))
}

func pickWinner() {
	// give 30 seconds each time to let the validators to propose the new block
	time.Sleep(30 * time.Second)
	stdout <- "Current Transactions Pool"
	stdout <- transactions
	stdout <- "Current Declined Transactions Pool"
	stdout <- declinedTrans
	stdout <- "current validator is: " + currentValidator
	transactions = checkBalance(transactions)

	if currentValidator == "" {
		currentValidator = validators[0].Validator
	}

	if len(transactions) > 0 {
		_, found := FindTransFromBlockchain(transactions[0])

		if found {
			declinedTrans = append(declinedTrans, transactions[0])
			transactions = remove(transactions, 0)
		}
	}

	if currentValidator == address {
		stdout <- "start generating Block: "
		//perfrom the check on whether the user has enough balance to make the transaction
		transactions = checkBalance(transactions)

		if len(Blockchain) > 1 {
			updateTransaction(Blockchain[len(Blockchain)-1])
		} else if len(Blockchain) == 1 {
			updateTransaction(Blockchain[0])
		}

		if len(transactions) > 0 {
			_, found := FindTransFromBlockchain(transactions[0])
			if found {
				declinedTrans = append(declinedTrans, transactions[0])
				transactions = remove(transactions, 0)
			}
		}

		if len(transactions) > 0 {
			//generate the block
			stdout <- "Generating Block"
			mutex.Lock()
			oldLastIndex := Blockchain[len(Blockchain)-1]
			mutex.Unlock()

			// change the validator to the next node
			currentValidator = getNextValidator(currentValidator).Validator

			newBlock, err := generateBlock(oldLastIndex, transactions[0], address, currentValidator)

			newblockJSON, err := json.Marshal(newBlock)
			if err != nil {
				log.Fatal(err)
			}

			stdout <- "New Proposed Block send out"
			broadcastBlockchain(string(newblockJSON), messageType.NewProposedBlock)
			//append the block to the blockchain

			Blockchain = append(Blockchain, newBlock)
			updateTransaction(newBlock)

			/*
				output, err := json.Marshal(Blockchain)
				if err != nil {
					log.Fatal(err)
				}



				//broadcast the lastest version blockchain
				//broadcastBlockchain(string(output), messageType.LatestBlockChain)
			*/

		}
	}

}

func getNextValidator(current string) Validator {
	for i, item := range validators {
		if item.Validator == current && i != (len(validators)-1) {
			return validators[i+1]
		}
	}
	return validators[0]
}

// Get balance
func getMyBalance(address string) int {
	var outBalance int
	var inBalance int
	for _, block := range Blockchain {
		if block.Transaction.From == address {
			outBalance += block.Transaction.Value

		}
		if block.Transaction.To == address {
			inBalance += block.Transaction.Value
		}
	}
	var balance Balance
	balance.Address = address
	balance.Balance = inBalance - outBalance
	return balance.Balance
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

// FindIndex takes a slice and looks for an element in it. If found it will
// return it's key, otherwise it will return -1 and a bool of false.
func FindIndex(slice []int, val int) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
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

//calculateBlockHash returns the hash of all block information
func calculateBlockHash(block Block) string {
	record := string(block.Index) + block.Timestamp + block.Transaction.Signature + block.PrevHash
	return calculateHash(record)
}

// generateBlock creates a new block using previous block's hash
func generateBlock(oldBlock Block, transaction Transaction, address string, nextValidator string) (Block, error) {
	var newBlock Block

	t := time.Now()

	newBlock.Index = oldBlock.Index + 1
	newBlock.Timestamp = t.String()
	newBlock.Transaction = transaction
	newBlock.PrevHash = oldBlock.Hash
	newBlock.Hash = calculateBlockHash(newBlock)
	newBlock.Validator = address
	newBlock.NextValidator = nextValidator

	return newBlock, nil
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

// FindTrans takes a slice and looks for an element in it. If found it will
// return it's key, otherwise it will return -1 and a bool of false.
func FindTransFromBlockchain(val Transaction) (int, bool) {
	for i, item := range Blockchain {
		if item.Transaction.Signature == val.Signature && item.Transaction.Date == val.Date && item.Transaction.From == val.From && item.Transaction.ID == val.ID && item.Transaction.Value == val.Value {
			return i, true
		}
	}
	return -1, false
}

// FindBlock takes a slice check if the block's validator has already made a proposal.
// If found it will return it's key, otherwise it will return -1 and a bool of false.
func FindBlock(slice []Block, b Block) (int, bool) {
	for i, item := range slice {
		if item.Validator == b.Validator {
			return i, true
		}
	}
	return -1, false
}

func updateTransaction(block Block) {
	counter := -1
	for i, tran := range transactions {
		if block.Transaction.Signature == tran.Signature {
			counter = i
			break
		}
	}
	if counter != -1 {
		declinedTrans = append(declinedTrans, transactions[counter])
		mutex.Lock()
		transactions = remove(transactions, counter)
		mutex.Unlock()
	}

}

func remove(transactions []Transaction, s int) []Transaction {
	return append(transactions[:s], transactions[s+1:]...)
}

//handle msg
func msgHandler(msgObj interface{}, topicObj interface{}) {
	go func() {

		msg := msgObj.([]byte)
		topic := topicObj.(string)

		m := unMarshalToString([]byte(msg))
		//fmt.Println(m.Type)
		//fmt.Println(m.Signature)
		switch topic {

		//receive from the broadcast
		case "/public/blockchain":
			switch m.Type {

			case messageType.NewTransaction:
				var transaction Transaction
				err := json.Unmarshal([]byte(m.Body), &transaction)
				if err != nil {
					log.Fatal(err)
				}
				if transaction.Signature == "" {
					return
				}
				_, found3 := FindTrans(declinedTrans, transaction)
				if found3 {
					output, err := json.Marshal(transaction)
					if err != nil {
						log.Fatal(err)
					}
					broadcastBlockchain(string(output), messageType.NewTransaction)
					return
				}
				_, found := FindTransFromBlockchain(transaction)
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
				err := json.Unmarshal([]byte(m.Body), &nextBlockChain)
				if err != nil {
					log.Fatal(err)
				}
				//fmt.Println(nextBlockChain)
				if len(nextBlockChain) > len(Blockchain) {
					//fmt.Println("reaplced by longer chain")
					//fmt.Println(nextBlockChain)
					//println("longest chain found, discard previous")
					//replace the current blockchain if the blockchain received is longer
					stdout <- "Replaced by a longer chain"

					Blockchain = nextBlockChain
					currentValidator = Blockchain[len(Blockchain)-1].NextValidator

				} else if len(nextBlockChain) == len(Blockchain) {
					currentValidator = Blockchain[len(Blockchain)-1].NextValidator

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
				err := json.Unmarshal([]byte(m.Body), &block)
				if err != nil {
					log.Fatal(err)
				}
				_, found := FindTransFromBlockchain(block.Transaction)
				if found {

				} else {
					updateTransaction(block)
					Blockchain = append(Blockchain, block)
					stdout <- "Block received, verified, and append to the Blockchain"
					stdout <- block
				}

				broadcastBlockchain(string(m.Body), messageType.NewProposedBlock)

			case messageType.NewJoinNode:
				var v Validator
				err := json.Unmarshal([]byte(m.Body), &v)
				if err != nil {
					log.Fatal(err)
				}

				found := false
				for _, item := range validators {
					if item.Validator == v.Validator {
						found = true
						break
					}
				}
				if !found && len(validators) < poolLimit {
					//if validators not in the validators pool, add it
					validators = append(validators, v)
					//sort the validator pool by timestamp of when they join in
					sort.Slice(validators, func(i, j int) bool {
						return validators[i].Timestamp < validators[j].Timestamp
					})
					stdout <- "Validators Pool"
					stdout <- validators
				}

				broadcastBlockchain(string(m.Body), messageType.NewJoinNode)

			}
		}
	}()

}

//在生成block前检查
