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

// MiniBlock represents a mini block within the blockchain.
type MiniBlock struct {
	Index        int            // Position of the mini-block in the blockchain
	Transactions []*Transaction  // Transactions included in the mini-block
	Hash         string         // Hash of the mini-block
	CurrentSize  int            // Current size of the mini-block in bytes
	IsFull       bool           // Indicates whether the mini-block is full
	Key          string         // Unique key for the mini-block
	MerkleRoot   string         // Merkle root for transactions integrity
	ValidatorSig string         // Signature of the validator proposing the mini-block
}

// SubBlock represents a medium block containing multiple mini-blocks.
type SubBlock struct {
	Index      int        // Position of the sub-block
	MiniBlocks []MiniBlock // Mini-blocks within this sub-block
	Hash       string     // Unique hash of the sub-block
	Key        string     // Unique key for the sub-block
	Validator  string     // Address of the validator proposing the sub-block
	IsFull     bool       // Indicates whether the sub-block is full
}

// MainBlock represents a large block containing multiple sub-blocks.
type MainBlock struct {
	Index     int        // Position of the main block
	SubBlocks []SubBlock // Sub-blocks within this main block
	Hash      string     // Unique hash of the main block
	Key       string     // Unique key for the main block
	Validator string     // Address of the validator proposing the main block
	Status    string     // Status of the block: Pending, Validated, Finalized
	IsFull    bool       // Indicates whether the main block is full
}

// Mutex to synchronize mining and consensus operations
var mutex sync.Mutex
var processedTransactions sync.Map        // Map to track processed transactions using sync.Map for concurrency safety
var availableMiniBlocks []MiniBlock       // List of available mini-blocks (not full)
var fullMiniBlocks []MiniBlock            // List of full mini-blocks

// updateMiniBlockLists updates the lists of available and full mini-blocks.
func updateMiniBlockLists(miniBlock *MiniBlock, selectedIndex int) {
	if miniBlock.IsFull {
		fullMiniBlocks = append(fullMiniBlocks, *miniBlock)
		availableMiniBlocks = append(availableMiniBlocks[:selectedIndex], availableMiniBlocks[selectedIndex+1:]...)
	} else {
		availableMiniBlocks[selectedIndex] = *miniBlock
	}
}

// MineTransaction processes transactions and assigns them to a mini-block in a sub-block.
func MineTransaction(transactions []*Transaction, mainBlock *MainBlock, validator string) (*MiniBlock, error) {
	txID := generateTransactionID(transactions)
	if _, loaded := processedTransactions.LoadOrStore(txID, true); loaded {
		return nil, errors.New("transaction already processed")
	}

	if len(availableMiniBlocks) == 0 {
		return nil, errors.New("no available mini-block to record transaction")
	}

	// Randomly select an available mini-block without locking the entire function
	selectedIndex := rand.Intn(len(availableMiniBlocks))
	mutex.Lock() // Lock only for critical section to update mini-block
	miniBlock := &availableMiniBlocks[selectedIndex]
	totalTransactionSize := calculateTransactionsSize(transactions)
	if miniBlock.CurrentSize+totalTransactionSize > MaxMiniBlockSize {
		mutex.Unlock()
		return nil, errors.New("mini-block size exceeded")
	}

	miniBlock.Transactions = append(miniBlock.Transactions, transactions...)
	miniBlock.CurrentSize += totalTransactionSize
	miniBlock.ValidatorSig = validator // Attach validator signature
	if miniBlock.CurrentSize == MaxMiniBlockSize {
		miniBlock.IsFull = true
	}
	updateMiniBlockLists(miniBlock, selectedIndex)
	mutex.Unlock()

	miniBlock.MerkleRoot = calculateMerkleRoot(transactions)
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
func NewTransaction(transactions []*Transaction, mainBlock *MainBlock, validator string) error {
	// Validate and sign transactions before adding them to the block
	for _, tx := range transactions {
		if !tx.Validate() {
			return errors.New("invalid transaction detected")
		}
	}

	miniBlock, err := MineTransaction(transactions, mainBlock, validator)
	if err != nil {
		return err
	}

	// Check if the main block is full after adding the mini-block
	if len(mainBlock.SubBlocks) == MaxMiniBlocks {
		mainBlock.IsFull = true
	}
	return nil
}

// ProcessTransactionsConcurrently processes transactions concurrently using unlimited Goroutines.
func ProcessTransactionsConcurrently(transactions []*Transaction, mainBlock *MainBlock, validator string) {
	var wg sync.WaitGroup
	transactionChannel := make(chan *Transaction, len(transactions))

	for _, tx := range transactions {
		transactionChannel <- tx
	}
	close(transactionChannel)

	for tx := range transactionChannel {
		wg.Add(1)
		go func(tx *Transaction) {
			defer wg.Done()
			_ = NewTransaction([]*Transaction{tx}, mainBlock, validator)
		}(tx)
	}
	wg.Wait()
}

// calculateMiniBlockHash generates the hash for a mini-block.
func calculateMiniBlockHash(index int, transactions []*Transaction) string {
	record := strconv.Itoa(index) + concatTransactions(transactions)
	hash := sha256.Sum256([]byte(record))
	return hex.EncodeToString(hash[:])
}

// calculateTransactionsSize calculates the total size of a list of transactions in bytes.
func calculateTransactionsSize(transactions []*Transaction) int {
	size := 0
	for _, tx := range transactions {
		size += len(tx.Sender) + len(tx.Receiver) + 8 + len(tx.Timestamp.String()) + len(tx.Signature)
	}
	return size
}

// concatTransactions concatenates all transactions into a single string.
func concatTransactions(transactions []*Transaction) string {
	var builder strings.Builder
	for _, tx := range transactions {
		builder.WriteString(tx.Sender)
		builder.WriteString(tx.Receiver)
		builder.WriteString(strconv.FormatFloat(tx.Amount, 'f', 2, 64))
		builder.WriteString(tx.Timestamp.String())
		builder.WriteString(tx.Signature)
	}
	return builder.String()
}

// generateTransactionID generates a unique ID for a transaction.
func generateTransactionID(transactions []*Transaction) string {
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

// calculateMerkleRoot calculates the Merkle root for a list of transactions.
func calculateMerkleRoot(transactions []*Transaction) string {
	if len(transactions) == 0 {
		return ""
	}
	var hashes []string
	for _, tx := range transactions {
		hashes = append(hashes, generateTransactionID([]*Transaction{tx}))
	}
	for len(hashes) > 1 {
		var newLevel []string
		for i := 0; i < len(hashes); i += 2 {
			if i+1 < len(hashes) {
				hash := sha256.Sum256([]byte(hashes[i] + hashes[i+1]))
				newLevel = append(newLevel, hex.EncodeToString(hash[:]))
			} else {
				newLevel = append(newLevel, hashes[i])
			}
		}
		hashes = newLevel
	}
	return hashes[0]
}
