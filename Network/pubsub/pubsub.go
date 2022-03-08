package pubsub

import (
	"fmt"
	"github.com/go-stomp/stomp"
	"sync"
)

var mutex = &sync.Mutex{}

type PubSub struct {
	Conn   *stomp.Conn
	subs   map[string]*stomp.Subscription
	closed bool
}

func Connect(serverAddr string) *PubSub {
	conn, err := stomp.Dial("tcp", serverAddr)

	if err != nil {
		fmt.Println("[pubsub] cannot connect to server: ", err.Error())
		return nil
	}

	ps := &PubSub{conn, make(map[string]*stomp.Subscription), false}
	fmt.Println("[pubsub] PubSub server connected....")

	return ps
}

func (ps *PubSub) Subscribe(topic string, msgHandler func(interface{}, interface{})) {
	if ps.closed {
		fmt.Println("[pubsub] PubSub is closed.")
		return
	}

	sub, err := ps.Conn.Subscribe(topic, stomp.AckAuto)
	if err != nil {
		fmt.Println("[pubsub] cannot subscribe to ", topic)
		fmt.Println(err.Error())
		return
	}

	ps.subs[topic] = sub
	//log.Println(len(ps.subs))
	fmt.Println("[pubsub] Subscribe to ", topic)
	go func() {
		for {
			mutex.Lock()
			msg := <-ps.subs[topic].C
			msgHandler(msg.Body, topic)
			mutex.Unlock()
		}
	}()
}

func (ps *PubSub) Unsubscribe(topic string) {
	if ps.closed {
		fmt.Println("[pubsub] PubSub is closed.")
		return
	}

	if _, exist := ps.subs[topic]; exist {
		err := ps.subs[topic].Unsubscribe()
		if err != nil {
			fmt.Println("[pubsub] Cannot unsubscribe to ", ps.subs[topic].Id(), " ", err.Error())
			return
		}
		delete(ps.subs, topic)
		// log.Println(len(ps.subs))
	} else {
		fmt.Println("[pubsub] Subscirption does not exist", topic)
		return
	}

	fmt.Println("[pubsub] Unsubscribe to ", topic)
	return
}

func (ps *PubSub) Publish(topic string, msg string) {
	if ps.closed {
		fmt.Println("[pubsub] PubSub is closed.")
		return
	}

	var err = ps.Conn.Send(topic, "text/plain", []byte(msg), nil)

	if err != nil {
		fmt.Println("failed to send to server", err)
		return
	}
	//fmt.Println("Publish finished topic : " + topic)
	//fmt.Println("Publish finished msg : " + msg)
}

func (ps *PubSub) Close() {

	for _, sub := range ps.subs {
		sub.Unsubscribe()
	}

	ps.Conn.Disconnect()

	ps.closed = true

	fmt.Println("[pubsub] PubSub closed.")
}
