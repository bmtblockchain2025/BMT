package blockchain

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
)

// Transaction represents a single transaction in the blockchain.
type Transaction struct {
	Sender    string  // Address of the sender
	Receiver  string  // Address of the receiver
	Amount    float64 // Amount being transferred
	Timestamp int64   // Unix timestamp of the transaction
	Hash      string  // Hash of the transaction
}

// NewTransaction creates a new transaction with given details.
func NewTransaction(sender, receiver string, amount float64, timestamp int64) (*Transaction, error) {
	// Validate inputs
	if sender == "" || receiver == "" {
		return nil, errors.New("sender and receiver addresses cannot be empty")
	}
	if amount <= 0 {
		return nil, errors.New("transaction amount must be positive")
	}

	// Create the transaction
	tx := &Transaction{
		Sender:    sender,
		Receiver:  receiver,
		Amount:    amount,
		Timestamp: timestamp,
	}

	// Calculate the hash
	tx.Hash = tx.CalculateHash()

	return tx, nil
}

// CalculateHash generates a hash for the transaction.
func (t *Transaction) CalculateHash() string {
	record := t.Sender + t.Receiver + string(t.Timestamp) + string(t.Amount)
	hash := sha256.Sum256([]byte(record))
	return hex.EncodeToString(hash[:])
}

// Validate checks if the transaction is valid.
func (t *Transaction) Validate() bool {
	return t.Hash == t.CalculateHash()
}
