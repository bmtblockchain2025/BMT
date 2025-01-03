package defi_test

import (
	"BMT-Blockchain/src/applications/defi"
	"testing"
	"time"
)

func TestFarmingPool(t *testing.T) {
	pool := defi.NewFarmingPool(0.01) // Reward rate: 0.01 BMT/second

	// Stake
	err := pool.Stake("Alice", 1000)
	if err != nil {
		t.Errorf("Failed to stake: %v", err)
	}
	if pool.GetUserStake("Alice") != 1000 {
		t.Errorf("Expected Alice's stake to be 1000, got %d", pool.GetUserStake("Alice"))
	}

	// Wait for rewards
	time.Sleep(2 * time.Second)

	// Claim rewards
	rewards, err := pool.ClaimRewards("Alice")
	if err != nil {
		t.Errorf("Failed to claim rewards: %v", err)
	}
	if rewards <= 0 {
		t.Error("Expected rewards to be greater than 0")
	}
}
