package gamefi

import (
	"errors"
	"sync"
	"time"
)

// StakedNFT represents an NFT that has been staked.
type StakedNFT struct {
	NFTID      string  // ID of the staked NFT
	Owner      string  // Owner of the NFT
	StakeTime  int64   // Timestamp when the NFT was staked
	RewardRate float64 // Reward rate for staking (e.g., 10 BMT/day)
}

// NFTStakingPool manages staking and rewards for NFTs.
type NFTStakingPool struct {
	StakedNFTs map[string]*StakedNFT // Mapping from NFT ID to StakedNFT
	mutex      sync.Mutex            // Mutex for thread safety
}

// NewNFTStakingPool initializes a new NFT staking pool.
func NewNFTStakingPool() *NFTStakingPool {
	return &NFTStakingPool{
		StakedNFTs: make(map[string]*StakedNFT),
	}
}

// StakeNFT stakes an NFT into the pool and starts earning rewards.
func (pool *NFTStakingPool) StakeNFT(owner, nftID string, rewardRate float64) error {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	if _, exists := pool.StakedNFTs[nftID]; exists {
		return errors.New("NFT is already staked")
	}

	pool.StakedNFTs[nftID] = &StakedNFT{
		NFTID:      nftID,
		Owner:      owner,
		StakeTime:  time.Now().Unix(),
		RewardRate: rewardRate,
	}
	return nil
}

// UnstakeNFT allows a user to withdraw their staked NFT.
func (pool *NFTStakingPool) UnstakeNFT(owner, nftID string) (float64, error) {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	stakedNFT, exists := pool.StakedNFTs[nftID]
	if !exists {
		return 0, errors.New("NFT is not staked")
	}
	if stakedNFT.Owner != owner {
		return 0, errors.New("only the owner can unstake this NFT")
	}

	// Calculate rewards
	rewards := pool.calculateRewards(stakedNFT)

	// Remove NFT from the pool
	delete(pool.StakedNFTs, nftID)

	return rewards, nil
}

// ClaimRewards allows a user to claim their staking rewards without unstaking.
func (pool *NFTStakingPool) ClaimRewards(owner, nftID string) (float64, error) {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	stakedNFT, exists := pool.StakedNFTs[nftID]
	if !exists {
		return 0, errors.New("NFT is not staked")
	}
	if stakedNFT.Owner != owner {
		return 0, errors.New("only the owner can claim rewards for this NFT")
	}

	// Calculate rewards
	rewards := pool.calculateRewards(stakedNFT)

	// Reset stake time
	stakedNFT.StakeTime = time.Now().Unix()

	return rewards, nil
}

// calculateRewards calculates staking rewards based on time staked.
func (pool *NFTStakingPool) calculateRewards(stakedNFT *StakedNFT) float64 {
	duration := time.Now().Unix() - stakedNFT.StakeTime
	return float64(duration) / 86400 * stakedNFT.RewardRate // Rewards per day
}
