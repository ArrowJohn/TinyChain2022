# Consensus Layer



### Required Package

github.com/davecgh/go-spew/spew

github.com/gorilla/mux

github.com/go-stomp/stomp



github.com/shipingchen/BRA/Network/gossip

github.com/shipingchen/BRA/Network/message

### Introduction

This is the folder that runs on a local environment. To run the node on cloud instance, please forward to the folder " Clould_ConsensusLayer"

To run each of the Consensus Layer algorithm, please:

1. Have the bootnode running for gossip protocol, if you are running bootnode on cloud, then in the config.json, change the strategy to your bootnode public IPv4 address. If you are running bootnode on your local machine, simply change it to "local"

2. Have the activemq running for pubsub. 

3. Note that unlike cloud implementation, the way we distinguish the different nodes is not through their public address but through different ports. 

4. Follow the readme.md in each of the folder to run the nodes.

5. You can change the addresses list in the config file to decide which user address will be granted tokens in the Genesis Node. Please take a look at the Genesis Block generation codes in each of the file.

6. You can also determine the first validator in the genesisBlock in each of the appointment Selection file. In order to do this, simply change the informations in the Genesis Block generation code.
