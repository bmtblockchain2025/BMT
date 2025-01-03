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
	TargetMiningTime   = 2                // Target time (in seconds) for mining a block
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
}

// SubBlock represents a medium block containing multiple mini-blocks.
type SubBlock struct {
	Index      int        // Position of the sub-block
	MiniBlocks []MiniBlock // Mini-blocks within this sub-block
	Hash       string     // Unique hash of the sub-block
}

// MainBlock represents a large block containing multiple sub-blocks.
type MainBlock struct {
	Index     int        // Position of the main block
	SubBlocks []SubBlock // Sub-blocks within this main block
	Hash      string     // Unique hash of the main block
}

// Mutex to synchronize mining and consensus operations
var mutex sync.Mutex
var previousMiningTime = TargetMiningTime // Initialize with target time
var processedTransactions sync.Map // Map to track processed transactions using sync.Map for concurrency safety

// ValidateTransactionWithMultipleNodes performs consensus using multiple staking nodes.
func ValidateTransactionWithMultipleNodes(transactions []Transaction) bool {
	if len(transactions) == 0 {
		return false
	}

	var voteResults sync.Map
	var wg sync.WaitGroup
	requiredVotes := 2 * len(TrustedValidators) / 3

	// Simulate parallel voting by validators
	for validator := range TrustedValidators {
		wg.Add(1)
		go func(validator string) {
			defer wg.Done()
			if rand.Float64() > 0.5 { // Simulate random vote
				voteResults.Store(validator, true)
			}
		}(validator)
	}
	wg.Wait()

	votes := 0
	voteResults.Range(func(_, _ interface{}) bool {
		votes++
		return true
	})

	return votes >= requiredVotes
}

// MineTransaction processes transactions and assigns them to a mini-block in a sub-block.
func MineTransaction(transactions []Transaction, mainBlock *MainBlock) (*MiniBlock, error) {
	txID := generateTransactionID(transactions)
	if _, loaded := processedTransactions.LoadOrStore(txID, true); loaded {
		return nil, errors.New("transaction already processed")
	}

	suffix := calculateDynamicDifficulty()
	for _, subBlock := range mainBlock.SubBlocks {
		for i, miniBlock := range subBlock.MiniBlocks {
			if miniBlock.IsFull {
				continue // Skip full mini-blocks
			}
			// Check if the mini-block has enough space
			totalTransactionSize := calculateTransactionsSize(transactions)
			if miniBlock.CurrentSize+totalTransactionSize <= MaxMiniBlockSize {
				mutex.Lock()
				mainBlock.SubBlocks[subBlock.Index].MiniBlocks[i].Transactions = append(miniBlock.Transactions, transactions...)
				mainBlock.SubBlocks[subBlock.Index].MiniBlocks[i].CurrentSize += totalTransactionSize
				if mainBlock.SubBlocks[subBlock.Index].MiniBlocks[i].CurrentSize == MaxMiniBlockSize {
					mainBlock.SubBlocks[subBlock.Index].MiniBlocks[i].IsFull = true
				}
				mutex.Unlock()

				for {
					mainBlock.SubBlocks[subBlock.Index].MiniBlocks[i].Hash = calculateMiniBlockHash(miniBlock.Index, miniBlock.Transactions)
					if isValidHash(mainBlock.SubBlocks[subBlock.Index].MiniBlocks[i].Hash, suffix) {
						break
					}
				}
				mainBlock.SubBlocks[subBlock.Index].Hash = calculateSubBlockHash(subBlock)
				mainBlock.Hash = calculateMainBlockHash(mainBlock)
				return &mainBlock.SubBlocks[subBlock.Index].MiniBlocks[i], nil
			}
		}
	}
	return nil, errors.New("no available mini-block to record transaction")
}

// NewTransaction handles the full lifecycle of a transaction.
func NewTransaction(transactions []Transaction, mainBlock *MainBlock) error {
	if !ValidateTransactionWithMultipleNodes(transactions) {
		return errors.New("transaction validation failed")
	}

	_, err := MineTransaction(transactions, mainBlock)
	if err != nil {
		return err
	}
	return nil
}

// ProcessTransactionsConcurrently processes transactions concurrently using Goroutines.
func ProcessTransactionsConcurrently(transactions []Transaction, mainBlock *MainBlock, numWorkers int) {
	var wg sync.WaitGroup
	transactionChannel := make(chan Transaction, len(transactions))
	startTime := time.Now() // Start measuring time

	for _, tx := range transactions {
		transactionChannel <- tx
	}
	close(transactionChannel)

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for tx := range transactionChannel {
				_ = NewTransaction([]Transaction{tx}, mainBlock)
			}
		}()
	}
	wg.Wait()

	duration := time.Since(startTime) // Calculate elapsed time
	fmt.Printf("Processed %d transactions in %v\n", len(transactions), duration)
}

// calculateMiniBlockHash generates the hash for a mini-block.
func calculateMiniBlockHash(index int, transactions []Transaction) string {
	record := strconv.Itoa(index) + concatTransactions(transactions)
	hash := sha256.Sum256([]byte(record))
	return hex.EncodeToString(hash[:])
}

// calculateSubBlockHash generates the hash for a sub-block.
func calculateSubBlockHash(subBlock SubBlock) string {
	record := strconv.Itoa(subBlock.Index)
	for _, miniBlock := range subBlock.MiniBlocks {
		record += miniBlock.Hash
	}
	hash := sha256.Sum256([]byte(record))
	return hex.EncodeToString(hash[:])
}

// calculateMainBlockHash generates the hash for a main block.
func calculateMainBlockHash(mainBlock *MainBlock) string {
	record := strconv.Itoa(mainBlock.Index)
	for _, subBlock := range mainBlock.SubBlocks {
		record += subBlock.Hash
	}
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
	targetTime := TargetMiningTime
	actualTime := previousMiningTime // Simulate actual time taken
	suffixLength := 2
	if actualTime > targetTime {
		suffixLength = 3
	}
	return fmt.Sprintf("%0*d", suffixLength, 0)
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
