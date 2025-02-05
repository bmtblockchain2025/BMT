package api

import (
	"BMT-Blockchain/src/blockchain"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

const (
	apiKey = "secure-api-key"
	rateLimit = 5
)

type API struct {
	Bridge       *blockchain.CrossChainBridge
	Oracle       *blockchain.OracleSystem
	Blockchain   *blockchain.Blockchain
	RequestCount map[string]int
	mutex        sync.Mutex
}

func NewAPI(bridge *blockchain.CrossChainBridge, oracle *blockchain.OracleSystem, blockchain *blockchain.Blockchain) *API {
	return &API{
		Bridge:       bridge,
		Oracle:       oracle,
		Blockchain:   blockchain,
		RequestCount: make(map[string]int),
	}
}

func (api *API) authenticate(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-API-KEY") != apiKey {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}

func (api *API) StartAPI(port string) {
	http.HandleFunc("/get-transactions", api.authenticate(api.GetTransactionsHandler))
	http.HandleFunc("/get-latest-blocks", api.authenticate(api.GetLatestBlocksHandler))
	http.HandleFunc("/node-status", api.authenticate(api.NodeStatusHandler))
	http.HandleFunc("/get-peers", api.authenticate(api.GetPeersHandler))
	http.HandleFunc("/get-block-transactions", api.authenticate(api.GetBlockTransactionsHandler))

	log.Printf("API Server running on port %s", port)
	http.ListenAndServe(":"+port, nil)
}

func (api *API) GetTransactionsHandler(w http.ResponseWriter, r *http.Request) {
	address := r.URL.Query().Get("address")
	if address == "" {
		http.Error(w, "Missing address parameter", http.StatusBadRequest)
		return
	}
	transactions := api.Blockchain.GetTransactionsByAddress(address)
	json.NewEncoder(w).Encode(transactions)
}

func (api *API) GetLatestBlocksHandler(w http.ResponseWriter, r *http.Request) {
	latestBlocks := api.Blockchain.GetLatestBlocks(10)
	json.NewEncoder(w).Encode(latestBlocks)
}

func (api *API) NodeStatusHandler(w http.ResponseWriter, r *http.Request) {
	status := map[string]interface{}{
		"latestBlock": api.Blockchain.GetLatestBlock().Index,
		"pendingTransactions": len(api.Blockchain.PendingTransactions),
		"connectedPeers": len(api.Blockchain.P2PNetwork.Peers),
	}
	json.NewEncoder(w).Encode(status)
}

func (api *API) GetPeersHandler(w http.ResponseWriter, r *http.Request) {
	peers := []string{}
	for peerID := range api.Blockchain.P2PNetwork.Peers {
		peers = append(peers, peerID)
	}
	json.NewEncoder(w).Encode(peers)
}

func (api *API) GetBlockTransactionsHandler(w http.ResponseWriter, r *http.Request) {
	blockIndex := r.URL.Query().Get("index")
	if blockIndex == "" {
		http.Error(w, "Missing index parameter", http.StatusBadRequest)
		return
	}
	block := api.Blockchain.GetBlockByIndex(blockIndex)
	if block == nil {
		http.Error(w, "Block not found", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(block.Transactions)
}
