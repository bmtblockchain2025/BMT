package main

import (
	"BMT-Blockchain/src/blockchain"
	"fmt"
)

func main() {
	// Initialize blockchain
	bmtChain := blockchain.NewBlockchain()

	// Display initial balances
	fmt.Println("Initial Balances:")
	fmt.Printf("System: %d BMT\n", bmtChain.Tokenomics.GetBalance("system"))

	// Mint new coins
	fmt.Println("\n--- Minting Coins ---")
	err := bmtChain.Tokenomics.MintCoins("Alice", 1000)
	if err != nil {
		fmt.Println("Minting Error:", err)
	} else {
		fmt.Println("Minted 1,000 BMT to Alice")
	}

	// Transfer coins
	fmt.Println("\n--- Transferring Coins ---")
	err = bmtChain.AddTransactionWithTokenomics("Alice", "Bob", 500)
	if err != nil {
		fmt.Println("Transaction Error:", err)
	} else {
		fmt.Println("Transferred 500 BMT from Alice to Bob")
	}

	// Validate blockchain
	fmt.Println("\n--- Validating Blockchain ---")
	if bmtChain.IsValid() {
		fmt.Println("Blockchain is valid.")
	} else {
		fmt.Println("Blockchain is invalid!")
	}

	// Print blockchain
	fmt.Println("\nBlockchain:")
	for _, block := range bmtChain.Chain {
		fmt.Printf("Index: %d\n", block.Index)
		fmt.Printf("Transactions: %v\n", block.Transactions)
		fmt.Printf("Previous Hash: %s\n", block.PreviousHash)
		fmt.Printf("Hash: %s\n", block.Hash)
		fmt.Println("===================================")
	}

	// Display final balances
	fmt.Println("\nFinal Balances:")
	fmt.Printf("Alice: %d BMT\n", bmtChain.Tokenomics.GetBalance("Alice"))
	fmt.Printf("Bob: %d BMT\n", bmtChain.Tokenomics.GetBalance("Bob"))
	fmt.Printf("System: %d BMT\n", bmtChain.Tokenomics.GetBalance("system"))
}
