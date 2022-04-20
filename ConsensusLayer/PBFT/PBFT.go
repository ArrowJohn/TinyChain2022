package main

import (
	"TinyChain/ConsensusLayer/General"
	"TinyChain/ConsensusLayer/PBFT/network"
	"flag"
)

func main() {
	readConfig()
	General.Connect()
	General.InitTransaction()
	p := flag.String("p", "", "port")
	flag.Parse()
	sPort := *p
	if sPort != "" {
		ServerPort = sPort
	}
	network.NewServer(ServerPort, PublicAdd)
}
