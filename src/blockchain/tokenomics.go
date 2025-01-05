package blockchain

import (
	"errors"
	"sync"
)

// Tokenomics defines the properties and logic of the BMT Coin.
type Tokenomics struct {
	TotalSupply      float64            // Total supply of BMT coins
	MaxSupply        float64            // Maximum supply allowed
	Balances         map[string]float64 // Mapping of addresses to balances
	StakingRewards   float64            // Accumulated staking rewards
	TransactionMutex sync.Mutex         // Mutex for thread-safe operations
}

// NewTokenomics initializes the tokenomics with total supply and max supply.
func NewTokenomics(totalSupply, maxSupply float64) *Tokenomics {
	return &Tokenomics{
		TotalSupply:    totalSupply,
		MaxSupply:      maxSupply,
		Balances:       make(map[string]float64),
		StakingRewards: 0,
	}
}

// Transfer handles the transfer of BMT coins between wallets.
func (t *Tokenomics) Transfer(from, to string, amount float64) error {
	t.TransactionMutex.Lock()
	defer t.TransactionMutex.Unlock()

	if amount <= 0 {
		return errors.New("transfer amount must be greater than zero")
	}

	// Check if the sender has enough balance
	if t.Balances[from] < amount {
		return errors.New("insufficient balance")
	}

	// Perform the transfer
	t.Balances[from] -= amount
	t.Balances[to] += amount

	return nil
}

// MintCoins adds new coins to a specified wallet (e.g., rewards or incentives).
func (t *Tokenomics) MintCoins(to string, amount float64) error {
	t.TransactionMutex.Lock()
	defer t.TransactionMutex.Unlock()

	if amount <= 0 {
		return errors.New("mint amount must be greater than zero")
	}

	// Ensure we do not exceed max supply
	if t.TotalSupply+amount > t.MaxSupply {
		return errors.New("minting exceeds max supply")
	}

	// Mint coins
	t.Balances[to] += amount
	t.TotalSupply += amount

	return nil
}

// BurnCoins permanently removes coins from circulation.
func (t *Tokenomics) BurnCoins(from string, amount float64) error {
	t.TransactionMutex.Lock()
	defer t.TransactionMutex.Unlock()

	if amount <= 0 {
		return errors.New("burn amount must be greater than zero")
	}

	if t.Balances[from] < amount {
		return errors.New("insufficient balance to burn")
	}

	t.Balances[from] -= amount
	t.TotalSupply -= amount

	return nil
}

// RewardMiner rewards the miner with a fixed amount of BMT for creating a new block.
func (t *Tokenomics) RewardMiner(minerAddress string, rewardAmount float64) error {
	t.TransactionMutex.Lock()
	defer t.TransactionMutex.Unlock()

	if rewardAmount <= 0 {
		return errors.New("reward amount must be greater than zero")
	}

	// Ensure the total supply does not exceed the max supply
	if t.TotalSupply+rewardAmount > t.MaxSupply {
		return errors.New("rewarding exceeds max supply")
	}

	// Mint the reward to the miner's address
	t.Balances[minerAddress] += rewardAmount
	t.TotalSupply += rewardAmount

	return nil
}

// DistributeStakingRewards distributes accumulated staking rewards to validators.
func (t *Tokenomics) DistributeStakingRewards(validators map[string]float64) error {
	t.TransactionMutex.Lock()
	defer t.TransactionMutex.Unlock()

	if len(validators) == 0 {
		return errors.New("no validators to distribute rewards")
	}

	totalStaked := 0.0
	for _, stake := range validators {
		totalStaked += stake
	}

	if totalStaked == 0 {
		return errors.New("total staked amount is zero")
	}

	for address, stake := range validators {
		reward := (stake / totalStaked) * t.StakingRewards
		t.Balances[address] += reward
	}

	// Reset staking rewards after distribution
	t.StakingRewards = 0
	return nil
}

// AccumulateStakingReward accumulates staking rewards from transaction fees.
func (t *Tokenomics) AccumulateStakingReward(amount float64) {
	t.TransactionMutex.Lock()
	defer t.TransactionMutex.Unlock()
	t.StakingRewards += amount
}

// GetBalance retrieves the balance of a specific wallet.
func (t *Tokenomics) GetBalance(address string) float64 {
	t.TransactionMutex.Lock()
	defer t.TransactionMutex.Unlock()
	return t.Balances[address]
}
