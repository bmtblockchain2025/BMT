package blockchain

import (
	"crypto/sha256"
	"encoding/hex"
	"time"
)

// Block defines the structure of a single blockchain block.
type Block struct {
	Index        int       // Position of the block in the chain
	Timestamp    time.Time // Time of block creation
	Transactions []string  // Transactions included in the block
	PreviousHash string    // Hash of the previous block
	Hash         string    // Hash of the current block
	Nonce        int       // Nonce used for proof-of-work or consensus
}

// CalculateHash generates a SHA-256 hash for the block.
func (b *Block) CalculateHash() string {
	record := string(b.Index) + b.Timestamp.String() + b.PreviousHash + concatTransactions(b.Transactions) + string(b.Nonce)
	hash := sha256.Sum256([]byte(record))
	return hex.EncodeToString(hash[:])
}

// concatTransactions concatenates all transactions into a single string.
func concatTransactions(transactions []string) string {
	result := ""
	for _, tx := range transactions {
		result += tx
	}
	return result
}

// NewBlock creates a new block with the given data.
func NewBlock(index int, transactions []string, previousHash string) *Block {
	block := &Block{
		Index:        index,
		Timestamp:    time.Now(),
		Transactions: transactions,
		PreviousHash: previousHash,
		Nonce:        0,
	}
	block.Hash = block.CalculateHash()
	return block
}
