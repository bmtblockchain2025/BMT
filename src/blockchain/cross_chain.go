package blockchain

import (
	"errors"
	"fmt"
	"sync"
)

// CrossChainBridge manages cross-chain transactions.
type CrossChainBridge struct {
	LockedTokens map[string]int64 // Mapping of addresses to locked token amounts
	mutex        sync.Mutex       // Mutex for thread safety
}

// NewCrossChainBridge initializes a new cross-chain bridge.
func NewCrossChainBridge() *CrossChainBridge {
	return &CrossChainBridge{
		LockedTokens: make(map[string]int64),
	}
}

// LockTokens locks tokens on the source chain.
func (bridge *CrossChainBridge) LockTokens(address string, amount int64) error {
	bridge.mutex.Lock()
	defer bridge.mutex.Unlock()

	if amount <= 0 {
		return errors.New("amount must be greater than zero")
	}

	bridge.LockedTokens[address] += amount
	fmt.Printf("Locked %d tokens for %s\n", amount, address)
	return nil
}

// MintTokens mints tokens on the destination chain.
func (bridge *CrossChainBridge) MintTokens(address string, amount int64) error {
	bridge.mutex.Lock()
	defer bridge.mutex.Unlock()

	if amount <= 0 {
		return errors.New("amount must be greater than zero")
	}

	// Simulate minting tokens (can be integrated with smart contracts on other chains)
	fmt.Printf("Minted %d tokens for %s on the destination chain\n", amount, address)
	return nil
}

// VerifyCrossChainTransaction verifies the validity of a cross-chain transaction.
func (bridge *CrossChainBridge) VerifyCrossChainTransaction(txID string) (bool, error) {
	// Simulate verification logic (e.g., check Merkle proof or external oracle)
	if txID == "" {
		return false, errors.New("invalid transaction ID")
	}
	fmt.Printf("Transaction %s verified successfully\n", txID)
	return true, nil
}

// UnlockTokens unlocks tokens on the source chain after verification.
func (bridge *CrossChainBridge) UnlockTokens(address string, amount int64) error {
	bridge.mutex.Lock()
	defer bridge.mutex.Unlock()

	if amount > bridge.LockedTokens[address] {
		return errors.New("insufficient locked tokens")
	}

	bridge.LockedTokens[address] -= amount
	fmt.Printf("Unlocked %d tokens for %s\n", amount, address)
	return nil
}

// VerifyCrossChainTransactionWithOracle integrates Oracle for cross-chain verification.
func (bridge *CrossChainBridge) VerifyCrossChainTransactionWithOracle(txID, blockchain string, oracle *OracleSystem) (bool, error) {
	data, err := oracle.FetchData(txID, blockchain)
	if err != nil {
		return false, fmt.Errorf("error fetching data from Oracle: %v", err)
	}

	// Use Oracle to validate the transaction
	valid, err := oracle.ValidateTransaction(txID)
	if err != nil {
		return false, fmt.Errorf("error validating transaction: %v", err)
	}

	fmt.Printf("Transaction %s verified with Oracle: %s\n", txID, data)
	return valid, nil
}
