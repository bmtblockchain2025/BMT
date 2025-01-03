// Package blockchain provides core blockchain structures and functions.
package blockchain

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"time"
)

// Constants for block size limits
const (
	MaxBlockSize       = 10 * 1024 * 1024 // 10 MB
	MaxSubBlockSize    = 1 * 1024 * 1024  // 1 MB
	MaxMiniBlockSize   = 100 * 1024       // 0.1 MB
	MaxTransactionsPerBlock = 10000       // Example max transactions per block
)

// Transaction represents a single transaction in the blockchain.
type Transaction struct {
	Sender    string  // Address of the sender
	Receiver  string  // Address of the receiver
	Amount    float64 // Amount being transferred
	Timestamp time.Time // Time of transaction creation
}

// Block defines the structure of a single blockchain block.
type Block struct {
	Index            int            // Position of the block in the chain
	Timestamp        time.Time      // Time of block creation
	Transactions     []Transaction  // Transactions included in the block
	PreviousHash     string         // Hash of the previous block
	Hash             string         // Hash of the current block
	Nonce            int            // Nonce used for consensus
	ProcessingTime   time.Duration  // Time taken to process the block
	TotalTransaction float64        // Total transaction amount in the block
	Reward           float64        // Reward for mining the block
	Message          string         // Custom message for the block (max 1000 characters, only for large blocks)
	Creator          string         // Name of the block creator
	SubBlocks        []SubBlock     // Sub-blocks within the block
}

// SubBlock represents a smaller block within a main block.
type SubBlock struct {
	Index        int            // Position of the sub-block in the main block
	Transactions []Transaction  // Transactions included in the sub-block
	Hash         string         // Hash of the sub-block
	MiniBlocks   []MiniBlock    // Mini-blocks within the sub-block
}

// MiniBlock represents a mini block within a sub-block.
type MiniBlock struct {
	Index        int            // Position of the mini-block in the sub-block
	Transactions []Transaction  // Transactions included in the mini-block
	Hash         string         // Hash of the mini-block
}

// CalculateHash generates a SHA-256 hash for the block.
func (b *Block) CalculateHash() string {
	record := strconv.Itoa(b.Index) + b.Timestamp.String() + b.PreviousHash + concatTransactions(b.Transactions) + strconv.Itoa(b.Nonce) + b.ProcessingTime.String() + strconv.FormatFloat(b.TotalTransaction, 'f', 2, 64) + strconv.FormatFloat(b.Reward, 'f', 2, 64) + b.Message + b.Creator
	hash := sha256.Sum256([]byte(record))
	return hex.EncodeToString(hash[:])
}

// concatTransactions concatenates all transactions into a single string.
func calculateBlockSize(transactions []Transaction) int {
	totalSize := 0
	for _, tx := range transactions {
		totalSize += len(tx.Sender) + len(tx.Receiver) + 8 + len(tx.Timestamp.String())
	}
	return totalSize
}

func concatTransactions(transactions []Transaction) string {
	result := ""
	for _, tx := range transactions {
		result += tx.Sender + tx.Receiver + strconv.FormatFloat(tx.Amount, 'f', 2, 64) + tx.Timestamp.String()
	}
	return result
}

// NewBlock creates a new block with the given data.
func NewBlock(index int, transactions []Transaction, previousHash string, message string, creator string, blockchain []Block, totalBlocks *int) (*Block, error) {
	if len(transactions) > MaxTransactionsPerBlock {
		return nil, errors.New("exceeds maximum transactions per block")
	}

	if calculateBlockSize(transactions) > MaxSubBlockSize && len(message) > 0 {
		return nil, errors.New("only large blocks can have messages")
	}

	if len(message) > 1000 {
		return nil, errors.New("message exceeds 1000 characters")
	}

	if !checkPreviousBlocksStorage(blockchain) {
		fmt.Println("Warning: previous blocks have not reached full storage capacity")
	}

	if !verifyNodeIntegrity(creator) {
		return nil, errors.New("node verification failed: potential fraud or untrusted AI detected")
	}

	startTime := time.Now()
	totalAmount := calculateTotalTransaction(transactions)
	reward := calculateBlockReward(totalAmount)

	block := &Block{
		Index:            index,
		Timestamp:        time.Now(),
		Transactions:     transactions,
		PreviousHash:     previousHash,
		Nonce:            0,
		TotalTransaction: totalAmount,
		Reward:           reward,
		Message:          message,
		Creator:          creator,
		SubBlocks:        createSubBlocks(transactions),
	}
	block.MineBlock()
	block.ProcessingTime = time.Since(startTime)
	*totalBlocks++
	lockBlock(block)
	return block, nil
}

// checkPreviousBlocksStorage verifies if previous blocks have reached their storage limit.
func checkPreviousBlocksStorage(blockchain []Block) bool {
	for _, block := range blockchain {
		if len(block.Transactions)*107 < MaxBlockSize {
			return false
		}
	}
	return true
}

// verifyNodeIntegrity checks if the node creating the block is legitimate.
func isTrustedNode(creator string) bool {
	trustedNodes := map[string]bool{
		"trusted-node-1": true,
		"trusted-node-2": true,
		"trusted-node-3": true,
	}
	return trustedNodes[creator]
}
	// Placeholder for actual node verification logic (e.g., cryptographic proof, reputation checks)
	if !isTrustedNode(creator) {
		return false
	}
	return true
}

// createSubBlocks divides the transactions into sub-blocks and mini-blocks.
func createSubBlocks(transactions []Transaction) []SubBlock {
	numSubBlocks := len(transactions) / (MaxSubBlockSize / 107) // Estimate number of sub-blocks
	subBlocks := make([]SubBlock, numSubBlocks)

	for i := 0; i < numSubBlocks; i++ {
		subTransactions := transactions[i*(MaxSubBlockSize/107) : (i+1)*(MaxSubBlockSize/107)]
		miniBlocks := createMiniBlocks(subTransactions)
		subBlocks[i] = SubBlock{
			Index:        i,
			Transactions: subTransactions,
			Hash:         calculateSubBlockHash(i, subTransactions),
			MiniBlocks:   miniBlocks,
		}
	}

	return subBlocks
}

// createMiniBlocks divides the sub-block transactions into mini-blocks.
func createMiniBlocks(transactions []Transaction) []MiniBlock {
	numMiniBlocks := len(transactions) / (MaxMiniBlockSize / 107) // Estimate number of mini-blocks
	miniBlocks := make([]MiniBlock, numMiniBlocks)

	for i := 0; i < numMiniBlocks; i++ {
		miniTransactions := transactions[i*(MaxMiniBlockSize/107) : (i+1)*(MaxMiniBlockSize/107)]
		miniBlocks[i] = MiniBlock{
			Index:        i,
			Transactions: miniTransactions,
			Hash:         calculateMiniBlockHash(i, miniTransactions),
		}
	}

	return miniBlocks
}

// calculateSubBlockHash generates the hash for a sub-block.
func calculateSubBlockHash(index int, transactions []Transaction) string {
	record := strconv.Itoa(index) + concatTransactions(transactions)
	hash := sha256.Sum256([]byte(record))
	return hex.EncodeToString(hash[:])
}

// calculateMiniBlockHash generates the hash for a mini-block.
func calculateMiniBlockHash(index int, transactions []Transaction) string {
	record := strconv.Itoa(index) + concatTransactions(transactions)
	hash := sha256.Sum256([]byte(record))
	return hex.EncodeToString(hash[:])
}

// calculateTotalTransaction calculates the total amount of transactions in the block.
func calculateTotalTransaction(transactions []Transaction) float64 {
	total := 0.0
	for _, tx := range transactions {
		total += tx.Amount
	}
	return total
}

// calculateBlockReward calculates the block reward based on total transaction amount.
func calculateBlockReward(totalTransaction float64) float64 {
	return 10.0 + 0.01*totalTransaction // Fixed base reward with 1% of total transaction amount
}

func lockBlock(block *Block) {
	// Prevent future modification of block reward and message
	block.Reward = block.Reward // Locked after creation
	block.Message = block.Message
}
	baseReward := 10.0 // Fixed base reward for mining
	txReward := 0.01 * totalTransaction // 1% of total transaction amount as additional reward
	return baseReward + txReward
}

// MineBlock performs a simple mining operation by ensuring the hash ends with a specific pattern.
func (b *Block) MineBlock() {
	suffix := calculateDynamicDifficulty() // Simple pattern for security with low energy consumption
	for {
		b.Hash = b.CalculateHash()
		if isValidHash(b.Hash, suffix) {
			break
		}
		b.Nonce++
	}
}

// isValidHash checks if a hash ends with the required suffix.
func calculateDynamicDifficulty() string {
	targetProcessingTime := 2 // Target processing time in seconds
	suffixLength := 2         // Initial suffix length
	if targetProcessingTime < 1 {
		suffixLength = 1
	}
	return fmt.Sprintf("%0*d", suffixLength, 0)
}

func isValidHash(hash string, suffix string) bool {
	return hash[len(hash)-len(suffix):] == suffix
}
