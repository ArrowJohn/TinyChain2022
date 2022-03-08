# Messgaing Performance Testing

The performance testing of Messgaing Layer is conducted in the AWS cloud computing environment.

You can find results from the data.zip.

Four metrics are used to evaluate the performance: 
* Received: It is the possibility that a node in the network will receive a message.
* Duplicated: It is the possibility of a message received by nodes that have already received the same message.
* Time (microseconds): It is the time of a message to reach a node in the network.
* Gossip (microseconds): It is the time of a message to complete the entire gossip procedure in the network. (only applies to the Gossip protocol)