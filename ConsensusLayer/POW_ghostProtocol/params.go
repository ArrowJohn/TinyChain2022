package main

import (
	"TinyChain/ConsensusLayer/General"
	"TinyChain/Network/gossip"
	"TinyChain/Network/message"
	"sync"
)

// Block represents each 'item' in the blockchain
type Block struct {
	General.BasicBlock
	MerkleRoot []byte `json:"-"`
	Difficulty int    `json:"-"`
	Nonce      string `json:"-"`
	Body       BlockBody
}

// POWTransaction represents each 'item' in the blockchain
type POWTransaction struct {
	General.Transaction
}

type POWGP struct {
}

//BlockBody represents the formal expansion of the transactions
type BlockBody struct {
	Transactions []POWTransaction
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

// Detail struct for some basic information of the chain
type Detail struct {
	BlockChainHeight  int `json:"blockChainHeight"`
	TransactionNumber int `json:"TransactionNumber"`
}

//the initial difficulty will be 0(1 actually, since will be increased when generate the fist block)
var difficulty = 0

// MaxDifficulty is the maximum difficulty you want to set
var MaxDifficulty int

// transIndex is number of transactions stored in a block
var transIndex int

var g *gossip.Gossip

var messageType = message.MessageType{Bootstrap: "Bootstrap", LatestBlockChain: "LatestBlockChain", CurrentWinner: "CurrentWinner", NewTransaction: "NewTransaction", NewProposedBlock: "NewProposedBlock", NewJoinNode: "NewJoinNode", TEST: "TEST"}

// Addresses : user addresses, the genesis node will grant them tokens
var Addresses []string

var address string

//PublicAdd public cloud address
var PublicAdd string

// PrivateAdd private cloud address
var PrivateAdd string

// Port for gp
var Port string

// ServerPort for Http
var ServerPort string

//Transactions : a slice where we store the Transaction
var transactions []POWTransaction

//declinedTrans : a slice where we store the  declined Transaction
var declinedTrans []POWTransaction

// Blockchain is a series of validated Blocks
var Blockchain []Block

// block
var mutex = &sync.Mutex{}

//all output will be made through this channel to clean up
var stdout = make(chan interface{})

// BootNodeAddress based on the strategy to settle the bootnode address
var BootNodeAddress string
