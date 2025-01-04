// Package blockchain provides core blockchain structures and functions.
package blockchain

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Constants for block size limits
const (
	MaxBlockSize       = 10 * 1024 * 1024 // 10 MB
	MaxSubBlockSize    = 1 * 1024 * 1024  // 1 MB
	MaxMiniBlockSize   = 100 * 1024       // 0.1 MB
	MaxTransactionsPerBlock = 10000       // Example max transactions per block
	TargetMiningTime   = 0.01             // Target time (in seconds) for mining
	MaxMiniBlocks      = 10               // Maximum number of mini-blocks per sub-block
)

// TrustedValidators represents a list of staking nodes.
var TrustedValidators = map[string]bool{
	"staking-node-1": true,
	"staking-node-2": true,
	"staking-node-3": true,
}

// Miners represents a list of mining nodes.
var Miners = []string{
	"miner-node-1",
	"miner-node-2",
	"miner-node-3",
}

// Transaction represents a single transaction in the blockchain.
type Transaction struct {
	Sender    string  // Address of the sender
	Receiver  string  // Address of the receiver
	Amount    float64 // Amount being transferred
	Timestamp time.Time // Time of transaction creation
}

// MiniBlock represents a mini block within the blockchain.
type MiniBlock struct {
	Index        int            // Position of the mini-block in the blockchain
	Transactions []Transaction  // Transactions included in the mini-block
	Hash         string         // Hash of the mini-block
	CurrentSize  int            // Current size of the mini-block in bytes
	IsFull       bool           // Indicates whether the mini-block is full
	Key          string         // Unique key for the mini-block
}

// SubBlock represents a medium block containing multiple mini-blocks.
type SubBlock struct {
	Index      int        // Position of the sub-block
	MiniBlocks []MiniBlock // Mini-blocks within this sub-block
	Hash       string     // Unique hash of the sub-block
	Key        string     // Unique key for the sub-block
}

// MainBlock represents a large block containing multiple sub-blocks.
type MainBlock struct {
	Index     int        // Position of the main block
	SubBlocks []SubBlock // Sub-blocks within this main block
	Hash      string     // Unique hash of the main block
	Key       string     // Unique key for the main block
}

// Mutex to synchronize mining and consensus operations
var mutex sync.Mutex
var previousMiningTime = TargetMiningTime // Initialize with target time
var processedTransactions sync.Map        // Map to track processed transactions using sync.Map for concurrency safety
var availableMiniBlocks []MiniBlock       // List of available mini-blocks (not full)
var fullMiniBlocks []MiniBlock            // List of full mini-blocks

// ValidateTransactionWithFastConsensus performs fast consensus using a small group of validators.
func ValidateTransactionWithFastConsensus(transactions []Transaction) bool {
	if len(transactions) == 0 {
		return false
	}
	requiredVotes := len(TrustedValidators) / 2 + 1
	votes := 0
	for validator := range TrustedValidators {
		if rand.Float64() > 0.5 { // Simulate random vote
			votes++
			if votes >= requiredVotes {
				return true
			}
		}
	}
	return false
}

// MineTransaction processes transactions and assigns them to a mini-block in a sub-block.
func MineTransaction(transactions []Transaction, mainBlock *MainBlock) (*MiniBlock, error) {
	txID := generateTransactionID(transactions)
	if _, loaded := processedTransactions.LoadOrStore(txID, true); loaded {
		return nil, errors.New("transaction already processed")
	}

	mutex.Lock()
	if len(availableMiniBlocks) == 0 {
		mutex.Unlock()
		return nil, errors.New("no available mini-block to record transaction")
	}

	// Randomly select an available mini-block
	selectedIndex := rand.Intn(len(availableMiniBlocks))
	miniBlock := &availableMiniBlocks[selectedIndex]
	totalTransactionSize := calculateTransactionsSize(transactions)
	if miniBlock.CurrentSize+totalTransactionSize > MaxMiniBlockSize {
		mutex.Unlock()
		return nil, errors.New("mini-block size exceeded")
	}

	miniBlock.Transactions = append(miniBlock.Transactions, transactions...)
	miniBlock.CurrentSize += totalTransactionSize
	if miniBlock.CurrentSize == MaxMiniBlockSize {
		miniBlock.IsFull = true
		fullMiniBlocks = append(fullMiniBlocks, *miniBlock)
		availableMiniBlocks = append(availableMiniBlocks[:selectedIndex], availableMiniBlocks[selectedIndex+1:]...)
	}
	mutex.Unlock()

	suffix := calculateDynamicDifficulty()
	for {
		miniBlock.Hash = calculateMiniBlockHash(miniBlock.Index, miniBlock.Transactions)
		if isValidHash(miniBlock.Hash, suffix) {
			break
		}
	}
	return miniBlock, nil
}

// NewTransaction handles the full lifecycle of a transaction.
func NewTransaction(transactions []Transaction, mainBlock *MainBlock) error {
	if !ValidateTransactionWithFastConsensus(transactions) {
		return errors.New("transaction validation failed")
	}

	_, err := MineTransaction(transactions, mainBlock)
	if err != nil {
		return err
	}
	return nil
}

// ProcessTransactionsConcurrently processes transactions concurrently using unlimited Goroutines.
func ProcessTransactionsConcurrently(transactions []Transaction, mainBlock *MainBlock) {
	var wg sync.WaitGroup
	transactionChannel := make(chan Transaction, len(transactions))

	for _, tx := range transactions {
		transactionChannel <- tx
	}
	close(transactionChannel)

	for tx := range transactionChannel {
		wg.Add(1)
		go func(tx Transaction) {
			defer wg.Done()
			_ = NewTransaction([]Transaction{tx}, mainBlock)
		}(tx)
	}
	wg.Wait()
}

// calculateMiniBlockHash generates the hash for a mini-block.
func calculateMiniBlockHash(index int, transactions []Transaction) string {
	record := strconv.Itoa(index) + concatTransactions(transactions)
	hash := sha256.Sum256([]byte(record))
	return hex.EncodeToString(hash[:])
}

// calculateTransactionsSize calculates the total size of a list of transactions in bytes.
func calculateTransactionsSize(transactions []Transaction) int {
	size := 0
	for _, tx := range transactions {
		size += len(tx.Sender) + len(tx.Receiver) + 8 + len(tx.Timestamp.String())
	}
	return size
}

// calculateDynamicDifficulty adjusts the difficulty based on the target mining time.
func calculateDynamicDifficulty() string {
	return "00" // Simulate fast mining with fixed difficulty
}

// concatTransactions concatenates all transactions into a single string.
func concatTransactions(transactions []Transaction) string {
	var builder strings.Builder
	for _, tx := range transactions {
		builder.WriteString(tx.Sender)
		builder.WriteString(tx.Receiver)
		builder.WriteString(strconv.FormatFloat(tx.Amount, 'f', 2, 64))
		builder.WriteString(tx.Timestamp.String())
	}
	return builder.String()
}

// generateTransactionID generates a unique ID for a transaction.
func generateTransactionID(transactions []Transaction) string {
	var builder strings.Builder
	for _, tx := range transactions {
		builder.WriteString(tx.Sender)
		builder.WriteString(tx.Receiver)
		builder.WriteString(strconv.FormatFloat(tx.Amount, 'f', 2, 64))
		builder.WriteString(tx.Timestamp.String())
	}
	hash := sha256.Sum256([]byte(builder.String()))
	return hex.EncodeToString(hash[:])
}
