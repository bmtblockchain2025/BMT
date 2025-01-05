package blockchain

import (
	"errors"
	"fmt"
	"sync"
)

// Blockchain represents the chain of blocks and tokenomics system.
type Blockchain struct {
	Chain        []*Block    // Slice of blocks
	Tokenomics   *Tokenomics // Tokenomics for managing BMT Coin
	mutex        sync.RWMutex  // RWMutex for concurrent access
	LockedBlocks int          // Number of locked blocks that won't be modified
}

// NewBlockchain initializes a new blockchain with tokenomics.
func NewBlockchain() *Blockchain {
	genesisBlock := NewBlock(0, []string{"Genesis Block"}, "0")
	tokenomics := NewTokenomics(8_000_000_000.0, 8_000_000_000.0)
	// Assign initial supply to the system wallet
	tokenomics.Balances["system"] = 8_000_000_000.0

	return &Blockchain{
		Chain:        []*Block{genesisBlock},
		Tokenomics:   tokenomics,
		LockedBlocks: 0,
	}
}

// AddBlock adds a new block to the chain with raw transaction data if the previous block is not locked.
func (bc *Blockchain) AddBlock(transactions []string) error {
	bc.mutex.Lock()
	defer bc.mutex.Unlock()

	if len(bc.Chain) <= bc.LockedBlocks {
		return errors.New("all blocks are locked, cannot add new block")
	}

	lastBlock := bc.GetLatestBlock()
	if lastBlock == nil {
		return errors.New("last block is nil")
	}

	newBlock := CreateBlock(lastBlock.Index+1, transactions, lastBlock.Hash)
	bc.Chain = append(bc.Chain, newBlock)
	return nil
}

// AddTransaction adds a new transaction to the blockchain after validation.
func (bc *Blockchain) AddTransaction(tx *Transaction) error {
	bc.mutex.Lock()
	defer bc.mutex.Unlock()

	// Validate the transaction
	if !tx.Validate() {
		return errors.New("invalid transaction")
	}

	// Verify the signature
	valid, err := VerifyTransactionSignature(tx.Sender, tx.Signature, tx.Hash)
	if err != nil || !valid {
		return errors.New("invalid transaction signature")
	}

	// Check sender's balance
	if bc.Tokenomics.GetBalance(tx.Sender) < tx.Amount+tx.Fee {
		return errors.New("insufficient balance")
	}

	// Update balances
	err = bc.Tokenomics.Transfer(tx.Sender, tx.Receiver, tx.Amount)
	if err != nil {
		return err
	}
	bc.Tokenomics.Transfer(tx.Sender, "miner", tx.Fee)

	// Add transaction to the latest block
	latestBlock := bc.GetLatestBlock()
	latestBlock.AddTransaction(tx)

	return nil
}

// AddTransactionBlock adds a block containing validated transactions to the blockchain if not locked.
func (bc *Blockchain) AddTransactionBlock(transactions []*Transaction) error {
	bc.mutex.Lock()
	defer bc.mutex.Unlock()

	if len(bc.Chain) <= bc.LockedBlocks {
		return errors.New("all blocks are locked, cannot add transaction block")
	}

	if len(transactions) == 0 {
		return errors.New("no transactions to add")
	}

	for _, tx := range transactions {
		if !tx.Validate() {
			return errors.New("invalid transaction detected")
		}
	}

	transactionData := ExtractTransactionData(transactions)
	lastBlock := bc.GetLatestBlock()
	if lastBlock == nil {
		return errors.New("last block is nil")
	}

	newBlock := CreateBlock(lastBlock.Index+1, transactionData, lastBlock.Hash)
	newBlock.MerkleRoot = CalculateMerkleRoot(transactions)
	bc.Chain = append(bc.Chain, newBlock)
	return nil
}

// IsValid checks if the blockchain is valid by verifying all blocks.
func (bc *Blockchain) IsValid() bool {
	bc.mutex.RLock()
	defer bc.mutex.RUnlock()

	return ValidateChain(bc.Chain)
}

// GetLatestBlock retrieves the last block in the blockchain.
func (bc *Blockchain) GetLatestBlock() *Block {
	return bc.Chain[len(bc.Chain)-1]
}

// LockBlocks locks the specified number of blocks to prevent modifications.
func (bc *Blockchain) LockBlocks(count int) error {
	bc.mutex.Lock()
	defer bc.mutex.Unlock()

	if count < 0 || count > len(bc.Chain) {
		return errors.New("invalid block count to lock")
	}
	bc.LockedBlocks = count
	return nil
}
