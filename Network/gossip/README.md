# Gossip

Gossip is one of the implementation of the TinyChain Messaging Layer. It implements the epidemic theory for disseminating information in the network. 

All Gossip API signature are available in [gossipAPIs](gossipAPIs.md)

## Using the API

### Instantiate a Gossip instance

To instantiate a Gossip instance, you have to provide 4 parameters: 
* pubHost - public host name of connection
* privHost - private host name of connection
* conn_port - port number of connection
* fanout - fanout value for gossiping

If you are running the instance on local environment, the pubHost and privHost would be same

An example of instantiate a Gossip instance on local environment:

```go
g := NewGossip("127.0.0.1", "127.0.0.1", "3001", 2)
```

An example of instantiate a Gossip instance on cloud environment (e.g. AWS EC2):

```go
g := NewGossip("54.79.123.123", "172.31.11.11", "3001", 2)
```

### Peer Management
If you do know the ip adress of peers, you can simply use AddPeer() and RemovePeer().

```go
g.AddPeer("54.79.111.111:8001")
g.RemovePeer("54.79.111.111:8001")
```

### Bootnode Request
To connect to the existing blockchain network, you nee to send Bootstrap request to an avaliable Bootnode. Please notes that the type of JSONMessage is Bootstrap, and you need to put the public address (where the gossip node is listen on) as the message Body.

```go
msg := JSONMessage{messageType.Bootstrap, g.PubAddr, time.Now().String(), "signature", 0}
g.SendDirect(msg, "localhost:3333") //address of bootnode
```

### Receiving messages
To receive message, you need to use the ReceiveLoop method and the NewMsg channel. Also you have to define your own way to continusely receives messages.

```go
g := NewGossip("127.0.0.1", "127.0.0.1", "3001", 2)

go g.ReceiveLoop()

go getMsg(g)

//e.g. example for loop to get messages 
func getMsg(g *Gossip) {
    for {
        select {
            case msg := <-g.NewMsg:
                myMsgHandler(msg) //your own message handler
        }
    }
}
```

### Sending messages
There are 2 ways to send messages in Gossip.
We are using the JSONMessage format message.

* **Broadcast** - broadcasting a message to the network. The message is send to randomly select members from the peer list. The number of member are set using the fanout value. e.g. fanout=2, it will select 2 peers randomly from all peers. For gossip purpose, the round value in message have to be 0.

A simple example of this in action is:
```go
msg := JSONMessage{messageType.TEST, "content", time.Now().String(), "signature", 0}

g.Broadcast(msg)
```

* **SendDirect** - sending a message direct to one peer/ip address. If you want to communicate with Bootnode, please use this method.

A simple example of this in action is:
```go
msg := JSONMessage{messageType.TEST, "content", time.Now().String(), "signature", 0} 

g.SendDirect(msg, "localhost:3333")
```

### Close
If you want to close the Gossip instance in run time, use the Close method to safely terminates all connection and channel.

A example of trying to use method after closing is like: 
```bash
>>> g.Close()
>>> g.Broadcast(msg)
[gossip] Gossip is closed.
```

### Purging
We are currently storing all messagies and bad peers (disconnected peers) in Gossip. Memory is extremely valuable and affecting our performance.
So if you are going to run the gossip for long period of time, here strongly suggesting to enable the Purging functionality.

* Uncomment line [275](https://github.com/shipingchen/BRA/blob/22f9a58e9dc7c44e1916fc73de26214acc875bc9/Network/gossip/gossip.go#L275), [276](https://github.com/shipingchen/BRA/blob/22f9a58e9dc7c44e1916fc73de26214acc875bc9/Network/gossip/gossip.go#L276) to start the auto purging.
```go
go g.purgePeers()
go g.msgStore.PurgeMsg()
```
* Purging clock is set at line [17](https://github.com/shipingchen/BRA/blob/22f9a58e9dc7c44e1916fc73de26214acc875bc9/Network/gossip/gossip.go#L17), [18](https://github.com/shipingchen/BRA/blob/22f9a58e9dc7c44e1916fc73de26214acc875bc9/Network/gossip/gossip.go#L18), feel free to change the time period.
```go 
var peerTicker = time.NewTicker(24 * time.Hour)
var msgTicker = time.NewTicker(30 * time.Minute)
```
