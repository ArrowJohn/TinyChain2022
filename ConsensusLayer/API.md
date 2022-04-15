# API

- (GET) /getBlockchainStatus
    - Url: [http://18.183.135.240:81/getBlockchainStatus](http://18.183.135.240:81/getBlockchainStatus)
    - Get some basic information of the blockchain.
    - Reponse:
    
    ```json
    {
    		"blockChainHeight":2,
    		"TransactionNumber":3
    }
    ```
    
- (GET) /getBlockchain
    - Get the whole blockChain.
    - Url: [http://18.183.135.240:81/getBlockchain](http://18.183.135.240:81/getBlockchain)
    - Reponse:
    
    ```json
    [
        {
            "Index": 0,
            "Timestamp": "1646124245",
            "Hash": "bac540339265e70e9da74ff4cda2798d1afb343ecc3d35ff36a5cd66f17a95df",
            "PrevHash": "",
            "Signature": "aeebad4a796fcc2e15dc4c6061b45ed9b373f26adfc798ca7d2d8cc58182718e",
            "Body": {
                "Transactions": [
                    {
                        "id": "0",
                        "date": "1646124245",
                        "from": "genesisTransaction",
                        "to": "db9cf6884b3983e488e4",
                        "value": 1000,
                        "signature": "7e89ac9d6c8dee1cd361ddaca3895d0a239491b7de30ba3519a7209d14f1eff5",
                        "hash": "138ddd4cd70fa5ee80a6d58d0f18888a57497ed4f130332320990baa9acd1137"
                    },
                    {
                        "id": "1",
                        "date": "1646124245",
                        "from": "genesisTransaction",
                        "to": "5c24967295a450bb96e3",
                        "value": 1000,
                        "signature": "4b192eaeec91f53b45ea4c90024042fc6ffe8b7e8b115f3f4a27f990095869e7",
                        "hash": "f8872b7c4e0e2a6b19fd1b2b3bd55abd36bd544cd77d5b2b4fb67af14bb940fc"
                    }
                ]
            }
        }
    ]
    ```
    
- (GET) /getPartBlockchain
    - Get part of the blockChain.
    - Url: [http://18.183.135.240:81/getPartBlockchain?Index=1&Size=10](http://18.183.135.240:81/getPartBlockchain?Index=1&Size=10)
    - Request
        - Index: Current scope size number
        - Size: Number of records displayed per scope
    - Reponse:
    
    ```json
    [
        {
            "Index": 0,
            "Timestamp": "1646124245",
            "Hash": "bac540339265e70e9da74ff4cda2798d1afb343ecc3d35ff36a5cd66f17a95df",
            "PrevHash": "",
            "Signature": "aeebad4a796fcc2e15dc4c6061b45ed9b373f26adfc798ca7d2d8cc58182718e",
            "Body": {
                "Transactions": [
                    {
                        "id": "0",
                        "date": "1646124245",
                        "from": "genesisTransaction",
                        "to": "db9cf6884b3983e488e4",
                        "value": 1000,
                        "signature": "7e89ac9d6c8dee1cd361ddaca3895d0a239491b7de30ba3519a7209d14f1eff5",
                        "hash": "138ddd4cd70fa5ee80a6d58d0f18888a57497ed4f130332320990baa9acd1137"
                    },
                    {
                        "id": "1",
                        "date": "1646124245",
                        "from": "genesisTransaction",
                        "to": "5c24967295a450bb96e3",
                        "value": 1000,
                        "signature": "4b192eaeec91f53b45ea4c90024042fc6ffe8b7e8b115f3f4a27f990095869e7",
                        "hash": "f8872b7c4e0e2a6b19fd1b2b3bd55abd36bd544cd77d5b2b4fb67af14bb940fc"
                    }
                ]
            }
        }
    ]
    ```
    
- (GET) /getLastestBlock
    - Get latest block of the current blockchain.
    - Url: [http://18.183.135.240:81/getLastestBlock](http://18.183.135.240:81/getLastestBlock)
    - Reponse:
    
    ```json
    {
        "Index": 1,
        "Timestamp": "1646124865",
        "Hash": "0bb4b30f9f7ea1f52586fc6814533cf25a681b12e55ec0852110642aad52445b",
        "PrevHash": "bac540339265e70e9da74ff4cda2798d1afb343ecc3d35ff36a5cd66f17a95df",
        "Signature": "",
        "Body": {
            "Transactions": [
                {
                    "id": "2",
                    "date": "1643081337",
                    "from": "db9cf6884b3983e488e4",
                    "to": "yy",
                    "value": 300,
                    "signature": "7ce2fa97a75da86c73d7cd71463468b91641caa1afcf26c0258b1186783cb589",
                    "hash": "e01a1159f5d3364a6724094a946f57c602488aad160bc7e48e8fd373d86a27d2"
                }
            ]
        }
    }
    ```
    
- (GET) /getAllTransactions
    - Get all the transactions that have been verified.
    - Url: [http://18.183.135.240:81/getAllTransactions](http://18.183.135.240:81/getAllTransactions)
    - Reponse:
    
    ```json
    [
        {
            "id": "0",
            "date": "1646124245",
            "from": "genesisTransaction",
            "to": "db9cf6884b3983e488e4",
            "value": 1000,
            "signature": "7e89ac9d6c8dee1cd361ddaca3895d0a239491b7de30ba3519a7209d14f1eff5",
            "hash": "138ddd4cd70fa5ee80a6d58d0f18888a57497ed4f130332320990baa9acd1137"
        },
        {
            "id": "1",
            "date": "1646124245",
            "from": "genesisTransaction",
            "to": "5c24967295a450bb96e3",
            "value": 1000,
            "signature": "4b192eaeec91f53b45ea4c90024042fc6ffe8b7e8b115f3f4a27f990095869e7",
            "hash": "f8872b7c4e0e2a6b19fd1b2b3bd55abd36bd544cd77d5b2b4fb67af14bb940fc"
        }
    ]
    ```
    
- (GET) /getTransactionByHash
    - Url: [http://18.183.135.240:81/getTransactionByHash?hash=138ddd4cd70fa5ee80a6d58d0f18888a57497ed4f130332320990baa9acd1137](http://18.183.135.240:81/getTransactionByHash?hash=138ddd4cd70fa5ee80a6d58d0f18888a57497ed4f130332320990baa9acd1137)
    - Response:
    
    ```json
    {
        "id": "0",
        "date": "1646124245",
        "from": "genesisTransaction",
        "to": "db9cf6884b3983e488e4",
        "value": 1000,
        "signature": "7e89ac9d6c8dee1cd361ddaca3895d0a239491b7de30ba3519a7209d14f1eff5",
        "hash": "138ddd4cd70fa5ee80a6d58d0f18888a57497ed4f130332320990baa9acd1137"
    }
    ```
    
- (GET) /getBlockByIndex
    - Url: [http://18.183.135.240:81/getBlockByIndex?index=0](http://18.183.135.240:81/getBlockByIndex?index=0)
    - Response:
    
    ```json
    {
        "Index": 0,
        "Timestamp": "1646124245",
        "Hash": "bac540339265e70e9da74ff4cda2798d1afb343ecc3d35ff36a5cd66f17a95df",
        "PrevHash": "",
        "Signature": "aeebad4a796fcc2e15dc4c6061b45ed9b373f26adfc798ca7d2d8cc58182718e",
        "Body": {
            "Transactions": [
                {
                    "id": "0",
                    "date": "1646124245",
                    "from": "genesisTransaction",
                    "to": "db9cf6884b3983e488e4",
                    "value": 1000,
                    "signature": "7e89ac9d6c8dee1cd361ddaca3895d0a239491b7de30ba3519a7209d14f1eff5",
                    "hash": "138ddd4cd70fa5ee80a6d58d0f18888a57497ed4f130332320990baa9acd1137"
                },
                {
                    "id": "1",
                    "date": "1646124245",
                    "from": "genesisTransaction",
                    "to": "5c24967295a450bb96e3",
                    "value": 1000,
                    "signature": "4b192eaeec91f53b45ea4c90024042fc6ffe8b7e8b115f3f4a27f990095869e7",
                    "hash": "f8872b7c4e0e2a6b19fd1b2b3bd55abd36bd544cd77d5b2b4fb67af14bb940fc"
                }
            ]
        }
    }
    ```
    
- (GET) /getTransaction/{user}
    - Get the transactions for the user specifies.
    - Url: [http://18.183.135.240:81/getTransaction/db9cf6884b3983e488e4](http://18.183.135.240:81/getTransaction/db9cf6884b3983e488e4)
    - Reponse:
    
    ```json
    [
        {
            "id": "0",
            "date": "1646124245",
            "from": "genesisTransaction",
            "to": "db9cf6884b3983e488e4",
            "value": 1000,
            "signature": "7e89ac9d6c8dee1cd361ddaca3895d0a239491b7de30ba3519a7209d14f1eff5",
            "hash": "138ddd4cd70fa5ee80a6d58d0f18888a57497ed4f130332320990baa9acd1137"
        }
    ]
    ```
    
- (GET)/ getCurrentTransactions
    - View transactions in recent days
    - Url: [http://18.183.135.240:81/getCurrentTransactions?address=db9cf6884b3983e488e4&days=2](http://18.183.135.240:81/getCurrentTransactions?address=db9cf6884b3983e488e4&days=2)
    - Request
        - address : user address
        - days : Number of days you want to query
    - Reponse:
    
    ```json
    [
        {
            "id": "0",
            "date": "1646712062",
            "from": "genesisTransaction",
            "to": "db9cf6884b3983e488e4",
            "value": 1000,
            "signature": "047e502c2966a74bcd4c02b75b31576b152ae42de1f19708ebc57454ddfec20d",
            "hash": "743f7aa2b90456eacf2585935e34ecab01a90c5091bb52131db3f018e858e436"
        }
    ]
    ```
    
- (GET) /getBalance/{user}
    - Get the balance of the specifed user.
    - Url: [http://18.183.135.240:81/getBalance/db9cf6884b3983e488e4](http://18.183.135.240:81/getBalance/db9cf6884b3983e488e4)
    - Response:
    
    ```json
    {
        "address": "db9cf6884b3983e488e4",
        "balance": 1000
    }
    ```
    
- (POST) /sendTransaction
    - Create a transaction
    - Url: [http://18.183.135.240:80/sendTransaction](http://localhost:3000/sendTransaction)
    - Request:
    
    ```json
    {
        "date":"1643081337",
        "from": "db9cf6884b3983e488e4",
        "signature": "7ce2fa97a75da86c73d7cd71463468b91641caa1afcf26c0258b1186783cb589",
        "to": "yy",
        "value": 300
    }
    ```
    
    - Reponse:
    
    ```json
    {
    		"code": 200,
    		"comment": "transaction created successfully"
    }
    ```

