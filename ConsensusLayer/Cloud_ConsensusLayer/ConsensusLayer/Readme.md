# Consensus Layer

The Consensus Layer use the messaging layer API to transmit the messages from  peer to peer. The Consensus itself contains two algorithms:
* 1. Proof-of-work

* 2. Proof-of-stake:

    * New Join Nodes Slection

    * Appointed Selection

The current configuration is made for AWS EC2 cloud implementation. To run on the cloud, you only need to configure the correct Public Address(IPv4) ,Private Address(IPv4) and the Port for RESTful APIs. For gossip protocol, you need both the public address and the private address. But for pubsub, only public address is needed(you can change the public address to any identification of such nodes). 

Make sure that bootnode is running for gossip Protocol and activemq is running for pubsub.

Unlike the local implementation, since the node runs on different cloud instances. The cloud implementation removes all the command line arguments to the config file. 


