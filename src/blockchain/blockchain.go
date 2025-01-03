package blockchain

// Blockchain represents the chain of blocks.
type Blockchain struct {
	Chain []*Block // Slice of blocks
}

// NewBlockchain initializes a new blockchain with a genesis block.
func NewBlockchain() *Blockchain {
	// Create the genesis block
	genesisBlock := NewBlock(0, []string{"Genesis Block"}, "0")
	return &Blockchain{
		Chain: []*Block{genesisBlock},
	}
}

// AddBlock adds a new block to the chain.
func (bc *Blockchain) AddBlock(transactions []string) {
	// Get the last block in the chain
	lastBlock := bc.Chain[len(bc.Chain)-1]

	// Create a new block with the last block's hash
	newBlock := NewBlock(lastBlock.Index+1, transactions, lastBlock.Hash)

	// Append the new block to the chain
	bc.Chain = append(bc.Chain, newBlock)
}

// IsValid checks if the blockchain is valid.
func (bc *Blockchain) IsValid() bool {
	for i := 1; i < len(bc.Chain); i++ {
		currentBlock := bc.Chain[i]
		previousBlock := bc.Chain[i-1]

		// Verify the hash of the current block
		if currentBlock.Hash != currentBlock.CalculateHash() {
			return false
		}

		// Verify the hash link between the current and previous blocks
		if currentBlock.PreviousHash != previousBlock.Hash {
			return false
		}
	}
	return true
}
