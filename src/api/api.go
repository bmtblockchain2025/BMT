package api

import (
	"BMT-Blockchain/src/blockchain"
	"encoding/json"
	"fmt"
	"net/http"
)

type API struct {
	Bridge *blockchain.CrossChainBridge
	Oracle *blockchain.OracleSystem
}

func NewAPI(bridge *blockchain.CrossChainBridge, oracle *blockchain.OracleSystem) *API {
	return &API{
		Bridge: bridge,
		Oracle: oracle,
	}
}

// StartAPI starts the RESTful API server
func (api *API) StartAPI(port string) {
	http.HandleFunc("/lock-tokens", api.LockTokensHandler)
	http.HandleFunc("/mint-tokens", api.MintTokensHandler)
	http.HandleFunc("/unlock-tokens", api.UnlockTokensHandler)
	http.HandleFunc("/verify-transaction", api.VerifyTransactionHandler)

	fmt.Printf("API Server running on port %s\n", port)
	http.ListenAndServe(":"+port, nil)
}

// LockTokensHandler locks tokens on the source chain
func (api *API) LockTokensHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		Address string `json:"address"`
		Amount  int64  `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := api.Bridge.LockTokens(request.Address, request.Amount)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Tokens locked successfully"})
}

// MintTokensHandler mints tokens on the destination chain
func (api *API) MintTokensHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		Address string `json:"address"`
		Amount  int64  `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := api.Bridge.MintTokens(request.Address, request.Amount)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Tokens minted successfully"})
}

// UnlockTokensHandler unlocks tokens on the source chain
func (api *API) UnlockTokensHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		Address string `json:"address"`
		Amount  int64  `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := api.Bridge.UnlockTokens(request.Address, request.Amount)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Tokens unlocked successfully"})
}

// VerifyTransactionHandler verifies a cross-chain transaction
func (api *API) VerifyTransactionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	txID := r.URL.Query().Get("txID")
	blockchain := r.URL.Query().Get("blockchain")

	if txID == "" || blockchain == "" {
		http.Error(w, "Missing txID or blockchain parameter", http.StatusBadRequest)
		return
	}

	valid, err := api.Oracle.ValidateTransactionOnChain(txID, blockchain)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]bool{"valid": valid})
}
