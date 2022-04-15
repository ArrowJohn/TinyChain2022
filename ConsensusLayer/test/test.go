package main

import (
	"TinyChain/ConsensusLayer/General"
)

func main() {
	General.Connect()
	General.InitTransaction()
	//General.InitTransaction()
	//General.QueryTransInBlock(1)
	//General.InitBlock(3)
}
