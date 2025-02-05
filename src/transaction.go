package blockchain

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strconv"
	"time"
)

// Transaction represents a single transaction in the blockchain.
type Transaction struct {
	Sender    string  // Address of the sender
	Receiver  string  // Address of the receiver
	Amount    float64 // Amount being transferred (supports up to 0.00000001 BMT)
	Fee       float64 // Transaction fee
	Timestamp int64   // Unix timestamp of the transaction
	Hash      string  // Hash of the transaction
	Signature string  // Digital signature of the transaction
	Anonymous bool    // Whether the transaction is anonymous
	ZKPProof  string  // Zero-Knowledge Proof for anonymous transactions
}

// NewTransaction creates a new transaction with given details.
func NewTransaction(sender, receiver string, amount, fee float64, anonymous bool) (*Transaction, error) {
	// Validate inputs
	if sender == "" || receiver == "" {
		return nil, errors.New("sender and receiver addresses cannot be empty")
	}
	if amount <= 0 {
		return nil, errors.New("transaction amount must be positive")
	}
	if fee < 0 {
		return nil, errors.New("transaction fee cannot be negative")
	}

	tx := &Transaction{
		Sender:    sender,
		Receiver:  receiver,
		Amount:    amount,
		Fee:       fee,
		Timestamp: time.Now().Unix(),
		Anonymous: anonymous,
	}

	tx.Hash = tx.CalculateHash()
	if anonymous {
		tx.ZKPProof = tx.GenerateZKP()
	}
	return tx, nil
}

// CalculateHash generates a hash for the transaction.
func (t *Transaction) CalculateHash() string {
	record := t.Sender + t.Receiver + strconv.FormatInt(t.Timestamp, 10) +
		strconv.FormatFloat(t.Amount, 'f', 8, 64) + strconv.FormatFloat(t.Fee, 'f', 8, 64)
	hash := sha256.Sum256([]byte(record))
	return hex.EncodeToString(hash[:])
}

// Validate checks if the transaction is valid.
func (t *Transaction) Validate() bool {
	if t.Anonymous {
		return t.ValidateZKP(t.ZKPProof)
	}
	return t.Hash == t.CalculateHash() && t.Amount > 0 && t.Fee >= 0
}

// SignTransaction signs the transaction using a given private key.
func (t *Transaction) SignTransaction(signature string) {
	t.Signature = signature
}

// VerifyTransactionSignature verifies the signature of the transaction.
func VerifyTransactionSignature(publicKey, signature, hash string) (bool, error) {
	// Reuse the VerifySignature function from wallet.go
	return VerifySignature(publicKey, signature, hash)
}

// Zero-Knowledge Proof integration (dummy example)
func (t *Transaction) GenerateZKP() string {
	hash := sha256.Sum256([]byte(t.Sender + t.Receiver + strconv.FormatFloat(t.Amount, 'f', 8, 64)))
	return hex.EncodeToString(hash[:])
}

func (t *Transaction) ValidateZKP(proof string) bool {
	expectedProof := t.GenerateZKP()
	return expectedProof == proof
}
