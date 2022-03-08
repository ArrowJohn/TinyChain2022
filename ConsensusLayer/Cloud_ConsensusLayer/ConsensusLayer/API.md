


* /getTransactions/{address}: get the transactions for the user specifies. METHOD:GET

>* request

>* respond array of JSON objects:  
[
{"date":"2020-10-06 09:20:37.7414039 +1100 AEDT m=+4.031270501","from":"Jane Doe","id":"3","signature":"asdadasd","to":"John","value":0},
{"date":"2020-10-06 09:20:37.7414039 +1100 AEDT m=+4.031270501","from":"John Doe","id":"3","signature":"asdadasd","to":"John","value":0}
]

* /getAllTransactions: get all the transactions that have been verified. METHOD:GET

>* request

>* respond array of JSON objects:  
[
{"date":"2020-10-06 09:20:37.7414039 +1100 AEDT m=+4.031270501","from":"Jane Doe","id":"3","signature":"asdadasd","to":"John","value":0},
{"date":"2020-10-06 09:20:37.7414039 +1100 AEDT m=+4.031270501","from":"John Doe","id":"3","signature":"asdadasd","to":"John","value":0}
]


* /sendTransaction: create a transaction. METHOD:POST

>* request JSON:  
{
    "date":"2020-10-06 09:20:37.7414039 +1100 AEDT m=+4.031270501",
    "from": "Jane Doe",
    "id": "3",
    "signature": "asdadasd",
    "to": "John",
    "value": 20
}

>*  respond Status Code  
{
    "statusCode":201
    "comment":"transaction create successfully"
}  
{
    "statusCode":500
    "comment":"something goes wrong please try again"
}

* /getLastestBlock: get latest block of the current blockchain. METHOD:GET

>* request:

>* respond:  
{
"index":0,
"timestamp":"2020-10-06 09:25:34.9769506 +1100 AEDT m=+3.046606701",
"transaction":{"date":"2015","from":"Jane Doe","id":"3","signature":"asdadasd","to":"John","value":20},
"hash":"Genesis transaction: 27/06/2020","prevHash":"6e340b9cffb37a989ca544e6bb780a2c78901d3fb33738768511a30617afa01d",
"validator":""
}

* /getBlockchain: get the whole Blockchain. METHOD:GET
>* request:

>* respond (array of JSON objects):  
[{"index":0,"timestamp":"2020-10-06 09:20:37.7414039 +1100 AEDT m=+4.031270501","transaction":{"date":0,"from":"Jane Doe","id":"3","signature":"asdadasd","to":"John","value":0},"hash":"Genesis transaction: 27/06/2020","prevHash":"6e340b9cffb37a989ca544e6bb780a2c78901d3fb33738768511a30617afa01d","validator":""},{"index":0,"timestamp":"2020-10-06 09:23:53.8945631 +1100 AEDT m=+200.184429701","transaction":{"date":1234,"from":"Jane Doe","id":"3","signature":"asdadasd","to":"John","value":2},"hash":"Genesis transaction: 27/06/2020","prevHash":"6e340b9cffb37a989ca544e6bb780a2c78901d3fb33738768511a30617afa01d","validator":""},{"index":0,"timestamp":"2020-10-06 09:25:28.2487209 +1100 AEDT m=+294.538587501","transaction":{"date":0,"from":"Jane Doe","id":"3","signature":"asdadasd","to":"John","value":20},"hash":"Genesis transaction: 27/06/2020","prevHash":"6e340b9cffb37a989ca544e6bb780a2c78901d3fb33738768511a30617afa01d","validator":""}]

* /getBalance/{address}: get the balance of the specifed user. METHOD:GET

>* request

>* respond (array of JSON objects):  
{
    "address": "ADDRESS",
    "balance": 20
}

* FOR POW only: /verifyTransaction/{blockIndex}: verify if such transaction has been verified by specified block. METHOD:GET
>* request(JSON)
{
    "date":"2015",
    "from": "aaaa",
    "id": "3",
    "signature": "xxxxxxxxx",
    "to": "xxxxx",
    "value": 4
}

>* respond with a message inidicate whether such transaction is verified by the block index.

* FOR POS only: /verifyTransaction/: verify if such transaction has been verified by the blockchain. METHOD:GET
>* request(JSON)
{
    "date":"2015",
    "from": "aaaa",
    "id": "3",
    "signature": "xxxxxxxxx",
    "to": "xxxxx",
    "value": 4
}

>* respond with a message inidicate whether such transaction is verified by the block index.


* /getDeclinedTransactions: get the declined transactions, the declined transaction is the one with invalid sender address, or the address with insufficient amounts. METHOD:POST

>* request

>* respond (array of JSON objects):  
