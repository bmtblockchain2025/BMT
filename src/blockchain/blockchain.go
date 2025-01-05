package blockchain

import (
	"errors"
	"fmt"
	"sync"
)

// Blockchain represents the chain of blocks and tokenomics system.
type Blockchain struct {
	Chain        []*MainBlock  // Slice of main blocks
	Tokenomics   *Tokenomics   // Tokenomics for managing BMT Coin
	Consensus    *Consensus    // Consensus mechanism
	mutex        sync.RWMutex  // RWMutex for concurrent access
	LockedBlocks int           // Number of locked blocks that won't be modified
}

// NewBlockchain initializes a new blockchain with tokenomics and consensus.
func NewBlockchain() *Blockchain {
	genesisBlock := &MainBlock{
		Index:     0,
		SubBlocks: []SubBlock{},
		Hash:      "0",
		Key:       "genesis-key",
		Validator: "system",
		Status:    "Finalized",
	}
	tokenomics := NewTokenomics(8_000_000_000.0, 8_000_000_000.0)
	// Assign initial supply to the system wallet
	tokenomics.Balances["system"] = 8_000_000_000.0

	return &Blockchain{
		Chain:        []*MainBlock{genesisBlock},
		Tokenomics:   tokenomics,
		Consensus:    NewConsensus(),
		LockedBlocks: 0,
	}
}

// AddBlock adds a new main block to the chain after consensus.
func (bc *Blockchain) AddBlock(newBlock *MainBlock) error {
	bc.mutex.Lock()
	defer bc.mutex.Unlock()

	if len(bc.Chain) <= bc.LockedBlocks {
		return errors.New("all blocks are locked, cannot add new block")
	}

	lastBlock := bc.GetLatestBlock()
	if lastBlock == nil {
		return errors.New("last block is nil")
	}

	newBlock.Index = lastBlock.Index + 1
	newBlock.Key = generateBlockKey(newBlock)

	// Select validators and reach consensus
	validators, err := bc.Consensus.SelectValidators(5)
	if err != nil {
		return err
	}

	agreed, err := bc.Consensus.ReachConsensus(validators)
	if err != nil {
		return err
	}

	if !agreed {
		return errors.New("consensus not reached, block rejected")
	}

	bc.Chain = append(bc.Chain, newBlock)
	return nil
}

// AddTransaction adds a new transaction to the blockchain after validation.
func (bc *Blockchain) AddTransaction(tx *Transaction, validator string) error {
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
	if latestBlock.IsFull {
		return errors.New("latest block is full, cannot add transaction")
	}

	subBlock := SubBlock{
		Index:      len(latestBlock.SubBlocks),
		MiniBlocks: []MiniBlock{},
		Validator:  validator,
	}
	latestBlock.SubBlocks = append(latestBlock.SubBlocks, subBlock)

	return nil
}

// GetLatestBlock retrieves the last block in the blockchain.
func (bc *Blockchain) GetLatestBlock() *MainBlock {
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
