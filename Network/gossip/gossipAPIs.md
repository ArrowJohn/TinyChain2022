# Gossip APIs

type Gossip is an implementation of the Gossip protocol

 ```go
type Gossip struct {
	fanout      int						//fanout value
	round       int						//expected maximum round for messages sent
	Port        string					//port number
	pubHost     string					//public host name
	privHost    string					//private host name
	PubAddr     string					//public ip
	PrivAddr    string 					//private ip
	alivePeers  []string 				//alive peers list
	deadPeers   map[string]time.Time	//dead peers list with disconncted time
	msgStore    MessageStore			//historical messages
	lock        sync.Mutex				//mutual exclusion lock
	NewMsg      chan JSONMessage		//new messages channel
	closed      bool					//is the instance closed
}
 ```

---

### Constant

`var peerTicker = time.NewTicker(24 * time.Hour)`

peerTicker is the timer to removes dead peers every 24 hours

`var msgTicker = time.NewTicker(30 * time.Minute)`

msgTicker is the timer to removes old messages every 30 minutes

----

### Methods

 ```go
func (g *Gossip) SetPubHost(pubHost string)
 ```

Update the public host name and public ip address

```go
func (g *Gossip) SetPrivHost(privHost string)
```

Update the private host name and private ip address

```go
func (g *Gossip) SetPort(port string)
```

Update the port number, reset private ip address and public ip address

```go
func calculateRound(n int) int
```

Calculates the maximum round based on the length of peers (i.e. size of network)

```go
func (g *Gossip) Close()
```

Terminates the Gossip instance and close new message channel

```go
func (g *Gossip) purgePeers()
```

Purges old peers using peerTikcer

```go
func (g *Gossip) removeDeadPeer(peer string)
```

Remove a specifc peer from dead peer list

```go
func (g *Gossip) AddPeer(ipAddress string)
```

Add new peer to alive peer list

```go
func (g *Gossip) RemovePeer(ipAddress string) error	
```

Remove a specific peer from alive peer list

```go
func (g *Gossip) GetAlivePeers() []string
```

Return the alive peer list

```go
func (g *Gossip) ReceiveLoop()
```

Creats the server using the private address,  then enters a endless loop to wait for new message coming in

```go
func (g *Gossip) gossiping(conn net.Conn)
```

The gossip flow for any new message coming in. It compare the new message with the internal messages histories, if the message has not been seen, the node would multicast the message to random peers. Otherwise, the node do nothing.  

```go
func (g *Gossip) handleBootstrapMsg(msg message.JSONMessage)	
```

The process flow for receives Bootstrap response from the Bootnode.

```go
func (g *Gossip) Broadcast(msg message.JSONMessage)
```

Boardcasting the message  to the network

```go
func (g *Gossip) SendDirect(msg message.JSONMessage, peer string) error 
```

Send message to a specific peer

```go
func (g *Gossip) sendData(msg message.JSONMessage, peer string)
```

Dial a specifc peer and send data direct to it

```go
func readData(conn net.Conn) (message.JSONMessage, error)
```

When a message arrives from connection, it should parse to the desire JSONMessage structure 

```go
func (g *Gossip) randomPeers() ([]string, error)
```

Get random peers based on fanout value

