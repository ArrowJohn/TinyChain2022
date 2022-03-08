package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"TinyChain/Network/message"

	"golang.org/x/crypto/bcrypt"
)

type MessageType struct {
	Bootstrap string
}

var messageType = MessageType{"Bootstrap"}

type Bootnode struct {
	nodeId string

	PubAddr string //ipAddress

	PrivAddr string

	pubHost string

	privHost string

	Port string

	peers []string //ipAddress of neighbours connected

	close chan bool

	lock sync.Mutex

	ticker *time.Ticker
}

func New(pubHost string, privHost string, conn_port string) *Bootnode {
	b := &Bootnode{
		pubHost:  pubHost,
		privHost: privHost,
		PubAddr:  getAddress(pubHost, conn_port),
		PrivAddr: getAddress(privHost, conn_port),
		Port:     conn_port,
		peers:    make([]string, 0), // initial capacity is 0
		close:    make(chan bool),
	}
	return b
}

func getAddress(host string, port string) string {
	return host + ":" + port
}

func (b *Bootnode) SetPubHost(pubHost string) {
	b.pubHost = pubHost
	b.PubAddr = getAddress(pubHost, b.Port)
}

func (b *Bootnode) SetPrivHost(privHost string) {
	b.privHost = privHost
	b.PrivAddr = getAddress(privHost, b.Port)
}

func (b *Bootnode) SetPort(port string) {
	b.Port = port
	b.PubAddr = getAddress(b.pubHost, b.Port)
	b.PrivAddr = getAddress(b.privHost, b.Port)
}

func (b *Bootnode) SetTimeTicker(heartBeat string) {
	if heartBeat == "" {
		heartBeat = "10m"
	}
	duration, _ := time.ParseDuration(heartBeat)
	b.ticker = time.NewTicker(duration)

	log.Printf("Setting: Ticker time is: %v\n", heartBeat)
}

func (b *Bootnode) SetNodeId(nodeId string) {
	if nodeId != "" {
		b.nodeId = nodeId
	} else {
		b.nodeId = generateRandHexId(16)
	}
	log.Printf("Setting: NodeId is: %v\n", b.nodeId)
}

func (b *Bootnode) SetPeers(peers []string) {
	if len(peers) > 0 {
		b.peers = peers
	}
	log.Printf("Setting: Predefined Peers are: %v\n", b.peers)
}

func (b *Bootnode) getNodeUrl() string {
	b.lock.Lock()
	defer b.lock.Unlock()

	id, err := hex.DecodeString(b.nodeId)
	if err != nil {
		panic(err)
	}
	//fmt.Println("decode bytes: ", id)

	h := hashAndSalt(id)

	return fmt.Sprintf("tinyNode://%v@%v", h, b.PubAddr)
}

func hashAndSalt(id []byte) string {

	hash, err := bcrypt.GenerateFromPassword(id, bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}

	return hex.EncodeToString(hash)
}

func (b *Bootnode) Host() {
	l, err := net.Listen("tcp", b.PrivAddr)
	if err != nil {
		log.Fatal("Error: " + err.Error())
	}

	defer l.Close()

	log.Printf("Bootnode is listening on %s!\n", l.Addr().String())

	fmt.Println(b.getNodeUrl())

	// go b.heartBeat()

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal("Error with accept: " + err.Error())
		}
		// log.Println("=====Address======")
		// log.Println(conn.RemoteAddr())
		// log.Println(conn.LocalAddr())
		// log.Println("=================")

		go b.handleEvent(conn)
	}
}

func (b *Bootnode) GetPeerCount() int {
	b.lock.Lock()
	defer b.lock.Unlock()
	return len(b.peers)
}

func (b *Bootnode) handleEvent(conn net.Conn) {
	// log.Println("Handle Event.....")
	msg, err := readData(conn)
	if err != nil {
		log.Println(err)
	}
	// log.Printf("Message received: %+v\n", msg)

	switch msg.Type {
		case messageType.Bootstrap:
			destination := msg.Body
			log.Printf("Received Boostrap Request from %s.\n", destination)
			b.AddPeer(msg.Body)
			err := b.sendPeers(destination)
			if err != nil {
				log.Println(err)
			} else {
				//b.AddPeer(msg.Body) //add address
			}
		default:
			log.Printf("Received a Non-Boostrap Request\n.")
			// log.Printf("Message received: %+v\n", msg)
	}
	conn.Close()
}

func (b *Bootnode) sendPeers(destination string) error {
	peers := b.getPeers()
	var content string

	if len(peers) > 0 {
		content = strings.Join(peers, ",")
	} else {
		content = ""
	}
	log.Printf("content %s\n", content)

	msg := message.JSONMessage{messageType.Bootstrap, content, b.nodeId, time.Now().String(), 0}
	err := sendMsg(msg, destination)
	if err != nil {
		return err
	}

	log.Printf("Sent response to %s\n", destination)
	return nil
}

func (b *Bootnode) RemovePeer(address string) {
	b.lock.Lock()
	defer b.lock.Unlock()

	for i, peer := range b.peers {
		if peer == address {
			b.peers = append(b.peers[:i], b.peers[i+1:]...)
			log.Printf("Removed one peer: %s\n", address)
		}
	}
}

func (b *Bootnode) heartBeat() {
	for range b.ticker.C {
		// fmt.Println("ticker")
		// fmt.Println(time.Now())
		for _, peer := range b.peers {
			b.checkLive(peer)
		}
	}
}

func (b *Bootnode) checkLive(peer string) {
	conn, err := net.DialTimeout("tcp", peer, 5*time.Second)
	if err != nil {
		b.RemovePeer(peer)
		return
	} else {
		conn.Close()
	}
}

func sendMsg(msg message.JSONMessage, destination string) error {
	conn, err := net.DialTimeout("tcp", destination, 5*time.Second)
	if err != nil {
		return err
	}
	defer conn.Close()

	// log.Println("Preparing data to send")
	dataToSend, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("Unable to parese data: %v\n", err)
	}
	for len(dataToSend) > 0 {
		n, err := conn.Write(dataToSend)
		if err != nil {
			return fmt.Errorf("Write to Peer failed: %v", err)
		}
		dataToSend = dataToSend[n:]
	}

	// log.Println("Successful send one message")
	return nil
}

func readData(conn net.Conn) (message.JSONMessage, error) {
	var msg message.JSONMessage
	err := json.NewDecoder(conn).Decode(&msg)
	// fmt.Printf( "------ READ MSG: %+v\n", msg)
	return msg, err
}

func (b *Bootnode) getPeers() []string {
	b.lock.Lock()
	defer b.lock.Unlock()

	if len(b.peers) == 0 {
		return make([]string, 0)
	}

	peers := make([]string, len(b.peers))
	copy(peers, b.peers)

	// log.Printf("Getting peers %v\n", peers)
	log.Printf("Getting peers list length %d.\n", len(peers))

	return peers
}

func (b *Bootnode) AddPeer(peer string) {
	b.lock.Lock()
	defer b.lock.Unlock()

	check := false
	for _, p := range b.peers {
		if peer == p {
			check = true
			log.Println("Peers already added")
		}
	}
	if !check {
		b.peers = append(b.peers, peer)
		log.Printf("Added new peer %s.\n", peer)
		log.Printf("Updated peers list length is %d.\n", len(b.peers))
	}
	// log.Printf("Current Peer count is: %d\n", len(b.peers)))
}

func generateRandHexId(n int) string {
	bytes := make([]byte, n)
	_, err := rand.Read(bytes)
	if err != nil {
		panic(err)
	}

	//fmt.Println(hash)

	h := hex.EncodeToString(bytes)

	//fmt.Println("NODE ID hex is:", h)
	return h
}

type Config struct {
	NodeId    string   `json:"nodeId"`
	Port      string   `json:"port"`
	PubHost   string   `json:"pubhost"`
	PrivHost  string   `json:"prihost"`
	Peers     []string `json:"peers"`
	HeartBeat string   `json:"heartbeat"`
}

func ReadConfigFile(filename string) Config {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Println("Please include the configuration file in command line argument, use flag -config=filename")
		log.Fatal(err)
	}

	var config Config

	err = json.Unmarshal(data, &config)

	if err != nil {
		log.Fatal(err)
	}

	//fmt.Printf("%+v\n", config)
	// fmt.Println(config.NodeId == "")
	return config
}

func main() {
	file, err := os.OpenFile("bootnode.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v\n", err)
	}

	defer file.Close()

	mw := io.MultiWriter(file, os.Stdout)
	log.SetOutput(mw)

	log.Println("---------- ! Log of bootnode ! ----------")
	configFile := flag.String("config", "", "path of configuration file")

	flag.Parse()
	//fmt.Println("configFile: ", *configFile)

	config := ReadConfigFile(*configFile)

	b := New(config.PubHost, config.PrivHost, config.Port)
	b.SetNodeId(config.NodeId)
	b.SetPeers(config.Peers)
	b.SetTimeTicker(config.HeartBeat)
	b.Host()
}
