package blockchain

import (
	"crypto/sha256"
	"container/heap"
	"encoding/hex"
	"errors"
	"sync"
	"strconv"
	"time"
	"math/rand"
)

// Constants for block size limits
const (
	MaxBlockSize       = 10 * 1024 * 1024 // 10 MB
	MaxSubBlockSize    = 1 * 1024 * 1024  // 1 MB
	MaxMiniBlockSize   = 100 * 1024       // 0.1 MB
	MaxTransactionsPerBlock = 10000       // Example max transactions per block
	TargetMiningTime   = 0.01             // Target time (in seconds) for mining
	MaxMiniBlocks      = 10               // Maximum number of mini-blocks per sub-block
	WorkerCount        = 10               // Number of workers in worker pool
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

// Priority Queue for mini-blocks
type MiniBlockHeap []MiniBlock

func (h MiniBlockHeap) Len() int           { return len(h) }
func (h MiniBlockHeap) Less(i, j int) bool { return h[i].CurrentSize < h[j].CurrentSize }
func (h MiniBlockHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h *MiniBlockHeap) Push(x interface{}) {
	*h = append(*h, x.(MiniBlock))
}
func (h *MiniBlockHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

// Mutex to synchronize mining and consensus operations
var (
	mutex          sync.Mutex
	processedTransactions sync.Map        // Map to track processed transactions using sync.Map for concurrency safety
	miniBlockQueue  MiniBlockHeap         // Priority queue for available mini-blocks
	fullMiniBlocks  []MiniBlock           // List of full mini-blocks
	wg             sync.WaitGroup
)

func init() {
	heap.Init(&miniBlockQueue)
}

// updateMiniBlockQueue updates the priority queue after processing a mini-block.
func updateMiniBlockQueue(miniBlock MiniBlock) {
	mutex.Lock()
	heap.Push(&miniBlockQueue, miniBlock)
	mutex.Unlock()
}

// MineTransaction processes transactions and assigns them to a mini-block in a sub-block.
func MineTransaction(transactions []*Transaction, mainBlock *MainBlock, validator string) (*MiniBlock, error) {
	txID := generateTransactionID(transactions)
	if _, loaded := processedTransactions.LoadOrStore(txID, true); loaded {
		return nil, errors.New("transaction already processed")
	}

	if miniBlockQueue.Len() == 0 {
		return nil, errors.New("no available mini-block to record transaction")
	}

	mutex.Lock()
	miniBlock := heap.Pop(&miniBlockQueue).(MiniBlock)
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
		fullMiniBlocks = append(fullMiniBlocks, miniBlock)
	} else {
		updateMiniBlockQueue(miniBlock)
	}
	mutex.Unlock()

	miniBlock.MerkleRoot = calculateMerkleRoot(transactions)
	suffix := calculateDynamicDifficulty()
	for {
		miniBlock.Hash = calculateMiniBlockHash(miniBlock.Index, miniBlock.Transactions)
		if isValidHash(miniBlock.Hash, suffix) {
			break
		}
	}
	return &miniBlock, nil
}

// ProcessTransactionsUsingWorkerPool processes transactions using a worker pool.
func ProcessTransactionsUsingWorkerPool(transactions []*Transaction, mainBlock *MainBlock, validator string) {
	transactionChannel := make(chan *Transaction, len(transactions))

	for i := 0; i < WorkerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for tx := range transactionChannel {
				_ = NewTransaction([]*Transaction{tx}, mainBlock, validator)
			}
		}()
	}

	for _, tx := range transactions {
		transactionChannel <- tx
	}
	close(transactionChannel)
	wg.Wait()
}

// NewTransaction handles the full lifecycle of a transaction.
func NewTransaction(transactions []*Transaction, mainBlock *MainBlock, validator string) error {
	for _, tx := range transactions {
		if !tx.Validate() {
			return errors.New("invalid transaction detected")
		}
	}

	miniBlock, err := MineTransaction(transactions, mainBlock, validator)
	if err != nil {
		return err
	}

	if len(mainBlock.SubBlocks) == MaxMiniBlocks {
		mainBlock.IsFull = true
	}
	return nil
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
