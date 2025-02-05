package blockchain

import (
	"errors"
	"fmt"
	"sync"
)

// OracleSystem fetches and validates external blockchain data.
type OracleSystem struct {
	Data  map[string]string // Mapping of transaction IDs to validation results
	mutex sync.Mutex        // Mutex for thread safety
}

// NewOracleSystem initializes a new Oracle system.
func NewOracleSystem() *OracleSystem {
	return &OracleSystem{
		Data: make(map[string]string),
	}
}

// FetchData fetches external data from a specified blockchain.
func (oracle *OracleSystem) FetchData(txID, blockchain string) (string, error) {
	oracle.mutex.Lock()
	defer oracle.mutex.Unlock()

	// Simulate fetching data from external blockchain
	if txID == "" || blockchain == "" {
		return "", errors.New("invalid transaction ID or blockchain")
	}
	data := fmt.Sprintf("Data for TX %s from %s", txID, blockchain)
	oracle.Data[txID] = data
	return data, nil
}

// ValidateTransaction validates a transaction using Oracle data.
func (oracle *OracleSystem) ValidateTransaction(txID string) (bool, error) {
	oracle.mutex.Lock()
	defer oracle.mutex.Unlock()

	data, exists := oracle.Data[txID]
	if !exists {
		return false, errors.New("transaction not found in Oracle system")
	}

	// Simulate validation logic
	if len(data) > 0 {
		return true, nil
	}
	return false, errors.New("transaction validation failed")
}
