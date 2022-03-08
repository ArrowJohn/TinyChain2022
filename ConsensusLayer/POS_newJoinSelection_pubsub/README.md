* please refer to config.json: 

>* PoolLimit : the limit of the pool size
>* addresses : addresses for the users

* to run the file please do:

go run newJoinSelection_pubsub.go -s 3000 -p 8891


go run newJoinSelection_pubsub.go -s 3001 -p 8892

......

-s is flag for the server port, you will need this to do the api calls
-p is the flag for p2p network port
