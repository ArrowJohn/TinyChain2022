package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

func readConfig() {
	jsonFile, err := os.Open("PBFT.json")
	if err != nil {
		jsonFile, _ = os.Open("ConsensusLayer/PBFT/PBFT.json")
	}
	byteValue, _ := ioutil.ReadAll(jsonFile)

	var config Config
	json.Unmarshal(byteValue, &config)
	for i := 0; i < len(config.Addresses); i++ {
		Addresses = append(Addresses, config.Addresses[i].Address)
	}
	transIndex = config.TransactionIndex
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
