package blockchain_test

import (
	"BMT-Blockchain/src/blockchain"
	"testing"
)

func TestBlockCreation(t *testing.T) {
	block := blockchain.NewBlock(1, []string{"Transaction 1", "Transaction 2"}, "0000")
	if block.Index != 1 {
		t.Errorf("Expected block index 1, got %d", block.Index)
	}
	if len(block.Transactions) != 2 {
		t.Errorf("Expected 2 transactions, got %d", len(block.Transactions))
	}
	if block.PreviousHash != "0000" {
		t.Errorf("Expected previous hash '0000', got %s", block.PreviousHash)
	}
}
