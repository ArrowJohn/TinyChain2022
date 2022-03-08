# Bootnode

Bootnode is peer sampling service that helps new node joins the network and peers management.

## How to Run
In order to host a Bootnode for your network, you have to prepare a JSON file for setting up some essential attributes of Bootnode.

There are total 6 attributes you can define at begining: 
* pubhost - public host name of connection
* privhost - private host name of connection
* port - port number of connection
* nodeId - bootnode identify
* peers - predefined peers 
* heartbeat - time period to check peers 

Out of all, (pubhost, privhost and port) are essential attributes you have to specify. Others are optional.
Check [easy_config.json](easy_config.json) and [full_config.json](full_config.json) as example.

Then simply specify the json file as config when you start the node.

```bash
go run bootnode.go -config={filepath}
```

You will see "Bootnode is listening on {IP Address}!" appears in terminal if the file path and attributes are correct.  
The Bootnode is meant to be run by itself after it hosted online. 

## Description of Functionalities
There are two functionlities of Bootnode

1. Bootsrtaping

The main purpose of Bootnode is help new gossip node join the current blockchain, which we call it "Bootstraping". 
The mechanism behind this is gossip node send "Bootstraping" request to hosted bootnode, then bootnode return peer list currently stored. Then the gossip node add each node from the list and requesting them to add itself to their peer list. 

2. Peer Management

Bootnode is responsible for storing all nodes in the network, so each node could have a full-view memebership. If you are going to host the bootnode for a long period of time, it is stongly recommending to enable the heartbeat check, where bootnode can send a hearbeat check to see if node is avaliable online, and remove any unavaliable node. Maximise the accuracy of nodes in the network.  
Just simply uncomment line [147](https://github.com/shipingchen/BRA/blob/aea06743083951e84c7075c2ec1c3cca92e31475/Network/bootnode/bootnode.go#L147), the node can run auto check at the heartbeat interval either the hearbeat value that set at configuration json file at begining, or 10 mins default. You can also change the time interval using [SetTimeTicker()](https://github.com/shipingchen/BRA/blob/aea06743083951e84c7075c2ec1c3cca92e31475/Network/bootnode/bootnode.go#L84)
```go
go b.heartBeat()
```


