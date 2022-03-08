* please refer to config.json: 

>* Firstvalidator : what node port will be the first validator
>* nodes : the node port that will appear in the p2p network
>* addresses : addresses for the users

* to run the file please do:

go run POS_appoint -s 3000 -p 8891 -nv 8892

go run POS_appoint -s 3001 -p 8892 -nv 8893

......


-s is flag for the server port, you will need this to do the api calls
-p is the flag for p2p network port
-nv is the node port that will become validator after the current node port does