package defi_test

import (
	"BMT-Blockchain/src/applications/defi"
	"testing"
)

func TestLendingPool(t *testing.T) {
	pool := defi.NewLendingPool(10000, 0.05)

	// Deposit
	err := pool.Deposit("Alice", 2000)
	if err != nil {
		t.Errorf("Failed to deposit: %v", err)
	}
	if pool.GetUserBalance("Alice") != 2000 {
		t.Errorf("Expected Alice's balance to be 2000, got %d", pool.GetUserBalance("Alice"))
	}

	// Borrow
	err = pool.Borrow("Alice", 800)
	if err != nil {
		t.Errorf("Failed to borrow: %v", err)
	}
	if pool.GetUserBorrowed("Alice") != 800 {
		t.Errorf("Expected Alice's borrowed amount to be 800, got %d", pool.GetUserBorrowed("Alice"))
	}

	// Repay
	err = pool.Repay("Alice", 400)
	if err != nil {
		t.Errorf("Failed to repay: %v", err)
	}
	if pool.GetUserBorrowed("Alice") != 400 {
		t.Errorf("Expected Alice's borrowed amount to be 400, got %d", pool.GetUserBorrowed("Alice"))
	}
}
