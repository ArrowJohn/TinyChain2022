# PubSub APIs

**type PubSub** is an implementation of Pub/Sub Messaging protocol

 ```go
type PubSub struct {
  Conn *stomp.Conn                    //the network connection
  subs map[string]*stomp.Subscription //the list of subscriptions map by topic name
  closed bool                         //is the instance closed
} 
 ```

----

### Methods

```go
func Connect(serverAddr string) *PubSub
```

Connect instantiate an object of PubSub. It takes one parameter, the network endpoint where the PubSub should listen on.  If connection establish successfully, it would instantatiate a new object of PubSub and return it. Otherwise it would return nil as indication of error.

```go
func (ps *PubSub) Subscribe(topic string, msgHandler func(interface{}, interface{}))
```

Subscribe funcation creates the new subscription on server. The function takes in a topic and msgHandler. Topic is the destination where messages will be sent. The msgHandler is the customised  function that will later handles the received messages from the created subscription.

```go
func (ps *PubSub) Unsubscribe(topic string)
```

Unsubcribes from the topic and closes the channel.

```go
func (ps *PubSub) Publish(topic string, msg string)
```

Publish sends a message to the topic destination in the server. 

