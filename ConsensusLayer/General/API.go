package General

import "net/http"

type API interface {
	GetBlockchainStatus(w http.ResponseWriter, r *http.Request)
	GetBlockchain(w http.ResponseWriter, r *http.Request)
	GetPartBlockchain(w http.ResponseWriter, r *http.Request)
	GetLatestBlock(w http.ResponseWriter, r *http.Request)
	GetBlockByIndex(w http.ResponseWriter, r *http.Request)

	GetAllTransactions(w http.ResponseWriter, r *http.Request)
	GetUserTransactions(w http.ResponseWriter, r *http.Request)
	GetBalance(w http.ResponseWriter, r *http.Request)
	CreateTransaction(w http.ResponseWriter, r *http.Request)
	GetTransByHash(w http.ResponseWriter, r *http.Request)
	GetCurrentTrans(w http.ResponseWriter, r *http.Request)
}
