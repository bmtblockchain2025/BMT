package blockchain

import "errors"

// Blockchain represents the chain of blocks and tokenomics system.
type Blockchain struct {
	Chain      []*Block    // Slice of blocks
	Tokenomics *Tokenomics // Tokenomics for managing BMT Coin
}

// NewBlockchain initializes a new blockchain with tokenomics.
func NewBlockchain() *Blockchain {
	genesisBlock := NewBlock(0, []string{"Genesis Block"}, "0")
	tokenomics := NewTokenomics(8_000_000_000, 10_000_000_000) // Total: 8B, Max: 10B

	// Assign initial supply to the system wallet
	tokenomics.Balances["system"] = 8_000_000_000

	return &Blockchain{
		Chain:      []*Block{genesisBlock},
		Tokenomics: tokenomics,
	}
}

// AddBlock adds a new block to the chain with raw transaction data.
func (bc *Blockchain) AddBlock(transactions []string) {
	lastBlock := bc.Chain[len(bc.Chain)-1]
	newBlock := NewBlock(lastBlock.Index+1, transactions, lastBlock.Hash)
	bc.Chain = append(bc.Chain, newBlock)
}

// AddTransactionBlock adds a block containing validated transactions to the blockchain.
func (bc *Blockchain) AddTransactionBlock(transactions []*Transaction) error {
	for _, tx := range transactions {
		if !tx.Validate() {
			return errors.New("invalid transaction detected")
		}
	}

	var transactionData []string
	for _, tx := range transactions {
		transactionData = append(transactionData, tx.Hash)
	}

	lastBlock := bc.Chain[len(bc.Chain)-1]
	newBlock := NewBlock(lastBlock.Index+1, transactionData, lastBlock.Hash)
	bc.Chain = append(bc.Chain, newBlock)

	return nil
}

// AddTransactionWithTokenomics adds a transaction to the blockchain and updates balances.
func (bc *Blockchain) AddTransactionWithTokenomics(from, to string, amount int64) error {
	err := bc.Tokenomics.Transfer(from, to, amount)
	if err != nil {
		return err
	}

	transaction := []string{from + " -> " + to + ": " + string(amount) + " BMT"}
	lastBlock := bc.Chain[len(bc.Chain)-1]
	newBlock := NewBlock(lastBlock.Index+1, transaction, lastBlock.Hash)
	bc.Chain = append(bc.Chain, newBlock)

	return nil
}

// IsValid checks if the blockchain is valid by verifying all blocks.
func (bc *Blockchain) IsValid() bool {
	for i := 1; i < len(bc.Chain); i++ {
		currentBlock := bc.Chain[i]
		previousBlock := bc.Chain[i-1]

		if currentBlock.Hash != currentBlock.CalculateHash() || currentBlock.PreviousHash != previousBlock.Hash {
			return false
		}
	}
	return true
}

// GetLatestBlock retrieves the last block in the blockchain.
func (bc *Blockchain) GetLatestBlock() *Block {
	return bc.Chain[len(bc.Chain)-1]
}
