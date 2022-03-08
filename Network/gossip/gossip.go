package gossip

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net"
	"strings"
	"sync"
	"time"

	"TinyChain/Network/message"
)

var peerTicker = time.NewTicker(24 * time.Hour)
var msgTicker = time.NewTicker(30 * time.Minute)

type MessageStore interface {
	Add(msg message.JSONMessage)
	Delete(idx int)
	Size() int
	CompareMsg(newMsg message.JSONMessage) bool
	PurgeMsg()
}

type messgaeStoreImpl struct {
	lock     sync.Mutex
	messages []message.JSONMessage
}

func (s *messgaeStoreImpl) Add(msg message.JSONMessage) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.messages = append(s.messages, msg)
}

func (s *messgaeStoreImpl) CompareMsg(newMsg message.JSONMessage) bool {
	s.lock.Lock()
	defer s.lock.Unlock()
	for _, msg := range s.messages {
		if msg.Time == newMsg.Time {
			return true
		}
	}
	return false
}

func (s *messgaeStoreImpl) Delete(i int) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.messages = append(s.messages[:i], s.messages[i+1:]...)
}

func (s *messgaeStoreImpl) PurgeMsg() {
	for range msgTicker.C {
		for i, msg := range s.messages {
			log.Println(i)
			layout := "2006-01-02 15:04:05.999999999 -0700 MST"
			timeString, err := time.Parse(layout, strings.Split(msg.Time, " m=")[0])
			if err == nil {
				if time.Since(timeString).Minutes() > 10 {
					s.Delete(i)
				}
			}
		}
	}
}

func (s *messgaeStoreImpl) Size() int {
	s.lock.Lock()
	defer s.lock.Unlock()
	return len(s.messages)
}

type Gossip struct {
	fanout int

	round int

	PubAddr string //ipAddress

	PrivAddr string

	pubHost string

	privHost string

	Port string

	lock sync.Mutex

	deadPeers map[string]time.Time

	alivePeers []string //ipAddress of peers connecting

	closed bool

	msgStore MessageStore

	NewMsg chan message.JSONMessage
}

func NewGossip(pubHost string, privHost string, conn_port string, fanout int) *Gossip {
	if fanout < 1 {
		fanout = 1
	}

	g := &Gossip{
		pubHost:    pubHost,
		privHost:   privHost,
		PubAddr:    getAddress(pubHost, conn_port),
		PrivAddr:   getAddress(privHost, conn_port),
		Port:       conn_port,
		alivePeers: make([]string, 0),
		deadPeers:  make(map[string]time.Time, 0),
		closed:     false,
		fanout:     fanout,
		NewMsg:     make(chan message.JSONMessage, 5),
		msgStore:   &messgaeStoreImpl{messages: make([]message.JSONMessage, 0)},
	}

	g.round = calculateRound(1)
	return g
}

func getAddress(host string, port string) string {
	return host + ":" + port
}

func (g *Gossip) SetPubHost(pubHost string) {
	g.pubHost = pubHost
	g.PubAddr = getAddress(pubHost, g.Port)
}

func (g *Gossip) SetPrivHost(privHost string) {
	g.privHost = privHost
	g.PrivAddr = getAddress(privHost, g.Port)
}

func (g *Gossip) SetPort(port string) {
	g.Port = port
	g.PubAddr = getAddress(g.pubHost, g.Port)
	g.PrivAddr = getAddress(g.privHost, g.Port)
}

func calculateRound(n int) int {
	x := math.Ceil(math.Log(float64(n + 1)))
	if x < 1 {
		x = 1
	}
	return int(x)
}

func (g *Gossip) Close() {
	g.lock.Lock()
	defer g.lock.Unlock()
	if !g.closed {
		close(g.NewMsg)
		g.closed = true
		fmt.Println("[gossip] Messaging Channle closed.")
	}
}

func (g *Gossip) IsClosed() bool {
	g.lock.Lock()
	defer g.lock.Unlock()
	return g.closed
}

func (g *Gossip) purgePeers() {
	for range peerTicker.C {
		for peer, timeStamp := range g.deadPeers {
			if time.Since(timeStamp).Hours() > 24 {
				g.removeDeadPeer(peer)
			}
		}
	}
}

func (g *Gossip) removeDeadPeer(peer string) {
	g.lock.Lock()
	defer g.lock.Unlock()
	delete(g.deadPeers, peer)
}

// validates the given ip adress
func isIpAddress(address string) bool {
	return strings.Count(address, ":") == 1
}

func (g *Gossip) AddPeer(ipAddress string) error {
	g.lock.Lock()
	defer g.lock.Unlock()

	if !isIpAddress(ipAddress) || ipAddress == g.PubAddr {
		fmt.Println("[gossip] AddPeer fails: not valid address")
		return fmt.Errorf("AddPeer fails: not valid address: %s\n", ipAddress)
	}

	for _, peer := range g.alivePeers {
		if peer == ipAddress {
			fmt.Printf("[gossip] Peer is already added: %s\n", ipAddress)
			return fmt.Errorf("Peer is already added: %s\n", ipAddress)
		}
	}

	for dPeer, _ := range g.deadPeers {
		if dPeer == ipAddress {
			delete(g.deadPeers, ipAddress)
			break
		}
	}

	g.alivePeers = append(g.alivePeers, ipAddress)

	g.round = calculateRound(len(g.alivePeers) + len(g.deadPeers))
	fmt.Printf("[gossip] Added new peer: %s\n", ipAddress)
	// log.Printf("[debug] new round is %d\n", g.round)
	// log.Printf("[debug] Added new peer %s\n", ipAddress)
	// log.Printf("[debug] alive peers: %v\n", g.alivePeers)
	// log.Printf("[debug] dead peers: %v\n", g.deadPeers)
	return nil
}

func (g *Gossip) RemovePeer(ipAddress string) error {
	g.lock.Lock()
	defer g.lock.Unlock()

	if g.closed {
		fmt.Println("[gossip] Gossip is closed.")
		return fmt.Errorf("[gossip] Gossip is closed.\n")
	}

	for i, peer := range g.alivePeers {
		if peer == ipAddress {
			g.alivePeers = append(g.alivePeers[:i], g.alivePeers[i+1:]...)
			g.deadPeers[ipAddress] = time.Now()

			g.round = calculateRound(len(g.alivePeers) + len(g.deadPeers))
			fmt.Printf("[gossip] Removed one peer: %s\n", ipAddress)
			// log.Printf("[debug] Remove one peer %s\n", ipAddress)
			// log.Printf("[debug] new round is %d\n", g.round)
			// log.Printf("[debug] alive peers: %v\n", g.alivePeers)
			// log.Printf("[debug] dead peers: %v\n", g.deadPeers)
			return nil
		}
	}
	return fmt.Errorf("Peer is not existes: %s\n", ipAddress)
}

func (g *Gossip) GetAlivePeers() []string {
	g.lock.Lock()
	defer g.lock.Unlock()

	peers := make([]string, len(g.alivePeers))
	copy(peers, g.alivePeers)

	return peers
}

func (g *Gossip) ReceiveLoop() {
	l, err := net.Listen("tcp", g.PrivAddr)
	if err != nil {
		log.Println(err)
		return
	}

	defer l.Close()
	g.closed = false

	log.Printf("Gossip is listening on %s\n", l.Addr().String())

	// go g.purgePeers()
	// go g.msgStore.PurgeMsg()

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Println("Error with accept: " + err.Error())
			return
		}

		go g.gossiping(conn)
	}
}

func (g *Gossip) handleBootstrapMsg(msg message.JSONMessage) {
	// log.Println("[debug] a new Bootstrap response received")
	peerList := strings.Split(msg.Body, ",")

	for _, peer := range peerList {
		err := g.AddPeer(peer)
		if err == nil {
			msg := message.JSONMessage{"NewJoinNode", g.PubAddr, time.Now().String(), "signature", 0}
			g.SendDirect(msg, peer)
		}
	}
}

func (g *Gossip) gossiping(conn net.Conn) {
	msg, err := readData(conn)

	if err != nil {
		log.Println("Error with reading data: ", err)
	} else {
		/*
			if msg.Type == "Bootstrap" {
				fmt.Printf("[gossip] Received a Bootstrap message.\n")
				g.handleBootstrapMsg(msg)
				// g.NewMsg <- msg
			} else if msg.Type == "NewJoinNode" {
				fmt.Printf("[gossip] Received a NewJoinNode message.\n")
				err := g.AddPeer(msg.Body)
				if err == nil {
					g.NewMsg <- msg
				}
			} else {
		*/
		isSeen := g.msgStore.CompareMsg(msg)
		if !isSeen {
			// log.Println("[debug] a new message received")
			fmt.Printf("[gossip] Received a new message.\n")

			g.msgStore.Add(msg)
			g.NewMsg <- msg
			// log.Printf("[debug] intial msg.Round: %d\n", msg.Round)
			for msg.Round < g.round {
				// log.Printf("[debug] currentRound: %d\n", msg.Round)
				chosenPeers, err := g.randomPeers()
				// log.Printf("[debug] chosenPeers: %v\n", chosenPeers)

				if err != nil {
					log.Println(err)
					return
				} else {
					msg.Round = msg.Round + 1
					for _, peer := range chosenPeers {
						g.sendData(msg, peer)
					}
				}
			}
		}
		//}
	}
	// log.Printf("[debug] gossip end ...\n")
}

func (g *Gossip) Broadcast(msg message.JSONMessage) {
	if g.closed {
		fmt.Println("[gossip] Gossip is closed.")
		return
	}
	fmt.Printf("[gossip] Broadcasting a new message\n")
	g.msgStore.Add(msg)
	// log.Printf("[debug] intial msg.Round: %d\n", msg.Round)
	for msg.Round < g.round {
		// log.Printf("[debug] currentRound: %d\n", msg.Round)
		chosenPeers, err := g.randomPeers()
		// log.Printf("[debug] chosenPeers: %v\n", chosenPeers)
		if err != nil {
			log.Println(err)
			return
		} else {
			msg.Round = msg.Round + 1
			for _, peer := range chosenPeers {
				g.sendData(msg, peer)
			}
		}
	}

}

func (g *Gossip) SendDirect(msg message.JSONMessage, peer string) error {
	if g.closed {
		fmt.Println("[gossip] Gossip is closed.")
		return fmt.Errorf("[gossip] Gossip is closed.\n")
	}

	fmt.Printf("[gossip] Sending a message to peer %s\n", peer)
	return g.sendData(msg, peer)
}

func (g *Gossip) sendData(msg message.JSONMessage, peer string) error {
	// fmt.Println("sendData")
	conn, err := net.DialTimeout("tcp", peer, 5*time.Second)
	// fmt.Println(err)
	if err != nil {
		g.RemovePeer(peer)
		// log.Printf("[debug] Peer %s is unavailable to gossip\n", peer)
		return fmt.Errorf("Peer %s is unavailable to gossip", peer)
	}

	dataToSend, err := json.Marshal(msg)
	if err != nil {
		log.Println(err)
		return err
	}

	for len(dataToSend) > 0 {
		n, err := conn.Write(dataToSend)
		if err != nil {
			log.Println(err)
			return err
		}
		dataToSend = dataToSend[n:]
	}
	// log.Printf("[debug] Sent message to Peer: %s\n", peer)
	conn.Close()
	return nil
}

func readData(conn net.Conn) (message.JSONMessage, error) {
	var msg message.JSONMessage
	err := json.NewDecoder(conn).Decode(&msg)

	// fmt.Println( "------ READ MSG -------")
	// fmt.Printf("%+v\n", msg)
	// fmt.Println("-------")
	conn.Close()
	return msg, err
}

func (g *Gossip) randomPeers() ([]string, error) {

	peers := make([]string, 0)
	length := len(g.alivePeers)
	count := g.fanout

	if length == 0 {
		return peers, fmt.Errorf("No peers to gossip with")
	} else if length < g.fanout {
		count = length
	}

	randIdx := rand.Perm(length)

	// log.Printf("[debug] Random Peers: fanout is %d, peers length is %d, final count is %d\n", g.fanout, length, count)
	// log.Printf("[debug] Random Peers index %v \n", randIdx)

	for i := 0; i < count; i++ {
		// fmt.Printf("[debug] Random Peers added: peers[%d] %s\n", randIdx[i], g.alivePeers[randIdx[i]])
		peers = append(peers, g.alivePeers[randIdx[i]])
	}

	return peers, nil
}
