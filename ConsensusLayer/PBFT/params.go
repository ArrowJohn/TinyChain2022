package main

import (
	"TinyChain/ConsensusLayer/General"
	"TinyChain/Network/gossip"
)

// Block represents each 'item' in the blockchain
type Block struct {
	General.BasicBlock
	Body BlockBody
}

type BlockBody struct {
	Transactions []PBFTTransaction
}

type PBFTTransaction struct {
	General.Transaction
}

type PBTF struct {
}

type Transaction struct {
	Date      string `json:"date"`
	From      string `json:"from"`
	ID        string `json:"id"`
	Signature string `json:"signature"`
	To        string `json:"to"`
	Value     int    `json:"value"`
}

var g *gossip.Gossip

var Blockchain []Block
var transIndex int
var Addresses []string
var BootNodeAddress string

//PublicAdd public cloud address
var PublicAdd string

// PrivateAdd private cloud address
var PrivateAdd string

// Port for gp
var Port string

type Address struct {
	Address string `json:"address"`
}

var ServerPort string

type Config struct {
	PublicAdd        string    `json:"PublicAdd"`
	PrivateAdd       string    `json:"PrivateAdd"`
	Port             string    `json:"Port"`
	ServerPort       string    `json:"ServerPort"`
	Strategy         string    `json:"Strategy"`
	Addresses        []Address `json:"addresses"`
	TransactionIndex int       `json:"transIndex"`
}
