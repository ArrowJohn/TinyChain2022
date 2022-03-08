# PubSub
PubSub is one of the implementation of the TinyChain Messaging Layer. It implements the Publisher/Subscribers pattern for disseminating information in the network. 
## How to Run
To use PubSub messaging, you have to run the Apache ActiveMQ first. It is the message broker server for queueing and sending messages in this implementation.
### 1. Set up Server

* Download <u>apache-activemq-5.15.11-bin.zip</u> from http://activemq.apache.org/components/classic/download/

* Go the folder where you unzip the package *e.g. \app\apache-activemq-5.15.11*

  1.1. Run 

  * window: `bin\activemq start`
  * macos: `./bin/activemq console`
  * The default username and password is **admin/admin**.  http://0.0.0.0:8161/admin

  1.2. stop (open another terminal tab in the project folder)

  * window: ``bin\activemq start``
  * macos: `./bin/activemq stop`

### 2. Set up your project folder

* GET **stomp** package
  
* In your go project folder, run`go get github.com/go-stomp/stomp` in terminal

* install **stomp**: go to each directory inside stomp and compile (`go install`) all source code, then it will generates *xxx.exe* and *xxx.a files* will go to *<u>bin</u>* and/or <u>*pkg*</u> automatically

## Using the API

### Instantiate a PubSub instance

To instantiate a PubSub instance, you have to provide 1 parameters: 
* serverAddr - ip adress following format "host:port"

An example of instantiate a PubSub instance on local environment:

```go
ps := pubsub.Connect("localhost:61613")
```

An example of instantiate a PubSub instance on cloud environment (e.g. AWS EC2):

```go
ps := pubsub.Connect("54.79.123.123:61613")
```

### Receiving messages
To receive incoming messages, you need to subscribe to the blockchain topic and provide your message handler function as parameter.

```go
ps.Subscribe("/topic/blockchain", myMsgHandler)
```

### Sending messages
To boardcast a message to the network, you need to publish a message to a topic.
We are using the JSONMessage format message.

```go
msg := JSONMessage{messageType.TEST, "content", time.Now().String(), "signature", 0}
ps.Publish("/topic/blockchain", msg)
```

### Close
If you want to close the PubSub instance in run time, use the Close method to safely terminates all connection and channel.

A example of tring to use method after closing is like: 
```bash
>>> ps.Close()
>>> ps.Publish("/topic/blockchain", msg)
[pubsub] PubSub is closed.
```