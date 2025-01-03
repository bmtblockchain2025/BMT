package defi

import (
	"errors"
	"sync"
	"time"
)

// FarmingPool represents a pool for yield farming.
type FarmingPool struct {
	TotalStaked   float64           // Total amount staked in the pool
	RewardRate    float64           // Reward rate per second (e.g., 0.01 BMT/s)
	Stakes        map[string]float64 // User stakes
	StakeTimes    map[string]int64  // Timestamps when users staked
	mutex         sync.Mutex        // Mutex for thread safety
}

// NewFarmingPool creates a new farming pool with a specified reward rate.
func NewFarmingPool(rewardRate float64) *FarmingPool {
	return &FarmingPool{
		RewardRate: rewardRate,
		Stakes:     make(map[string]float64),
		StakeTimes: make(map[string]int64),
	}
}

// Stake allows a user to stake tokens into the farming pool.
func (fp *FarmingPool) Stake(user string, amount float64) error {
	fp.mutex.Lock()
	defer fp.mutex.Unlock()

	if amount <= 0 {
		return errors.New("stake amount must be greater than zero")
	}

	// Calculate pending rewards before updating stake
	if _, exists := fp.StakeTimes[user]; exists {
		rewards := fp.calculateRewards(user)
		fp.TotalStaked += rewards
	}

	fp.TotalStaked += amount
	fp.Stakes[user] += amount
	fp.StakeTimes[user] = time.Now().Unix()
	return nil
}

// Unstake allows a user to withdraw staked tokens from the pool.
func (fp *FarmingPool) Unstake(user string, amount float64) (float64, error) {
	fp.mutex.Lock()
	defer fp.mutex.Unlock()

	if amount <= 0 {
		return 0, errors.New("unstake amount must be greater than zero")
	}
	if amount > fp.Stakes[user] {
		return 0, errors.New("unstake amount exceeds staked amount")
	}

	// Calculate pending rewards
	rewards := fp.calculateRewards(user)

	// Update staking data
	fp.TotalStaked -= amount
	fp.Stakes[user] -= amount
	delete(fp.StakeTimes, user)

	return rewards, nil
}

// ClaimRewards allows a user to claim their pending rewards.
func (fp *FarmingPool) ClaimRewards(user string) (float64, error) {
	fp.mutex.Lock()
	defer fp.mutex.Unlock()

	if _, exists := fp.StakeTimes[user]; !exists {
		return 0, errors.New("no active stake found for user")
	}

	rewards := fp.calculateRewards(user)
	fp.StakeTimes[user] = time.Now().Unix() // Reset stake time
	return rewards, nil
}

// calculateRewards calculates the rewards for a user based on staking duration.
func (fp *FarmingPool) calculateRewards(user string) float64 {
	stakeTime := fp.StakeTimes[user]
	duration := time.Now().Unix() - stakeTime
	return float64(fp.Stakes[user]) * float64(duration) * fp.RewardRate
}

// GetUserStake retrieves the user's staked amount.
func (fp *FarmingPool) GetUserStake(user string) float64 {
	fp.mutex.Lock()
	defer fp.mutex.Unlock()
	return fp.Stakes[user]
}
