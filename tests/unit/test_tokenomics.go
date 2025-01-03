package blockchain_test

import (
	"BMT-Blockchain/src/blockchain"
	"testing"
)

func TestMintCoins(t *testing.T) {
	tokenomics := blockchain.NewTokenomics(8_000_000_000, 10_000_000_000)
	err := tokenomics.MintCoins("Alice", 1000)
	if err != nil {
		t.Errorf("Failed to mint coins: %v", err)
	}
	if tokenomics.GetBalance("Alice") != 1000 {
		t.Errorf("Expected Alice's balance to be 1000, got %d", tokenomics.GetBalance("Alice"))
	}
}

func TestTransferCoins(t *testing.T) {
	tokenomics := blockchain.NewTokenomics(8_000_000_000, 10_000_000_000)
	tokenomics.MintCoins("Alice", 1000)
	err := tokenomics.Transfer("Alice", "Bob", 500)
	if err != nil {
		t.Errorf("Failed to transfer coins: %v", err)
	}
	if tokenomics.GetBalance("Alice") != 500 {
		t.Errorf("Expected Alice's balance to be 500, got %d", tokenomics.GetBalance("Alice"))
	}
	if tokenomics.GetBalance("Bob") != 500 {
		t.Errorf("Expected Bob's balance to be 500, got %d", tokenomics.GetBalance("Bob"))
	}
}
