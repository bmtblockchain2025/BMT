package blockchain

import (
	"errors"
	"math/rand"
	"sync"
	"time"
)

// Validator represents a staking node participating in the consensus.
type Validator struct {
	Address   string
	Stake     float64 // Amount of BMT staked
	IsTrusted bool    // Whether the validator is trusted
	Slashable bool    // Indicates if the validator can be slashed
}

// Consensus represents the consensus mechanism in the blockchain.
type Consensus struct {
	Validators     map[string]*Validator // List of validators
	StakingPool    float64               // Total staking pool
	mutex          sync.Mutex            // Mutex for thread safety
	ApprovalQuorum float64               // Percentage required for consensus
}

// NewConsensus initializes a new consensus mechanism.
func NewConsensus() *Consensus {
	return &Consensus{
		Validators:     make(map[string]*Validator),
		StakingPool:    0,
		ApprovalQuorum: 0.66, // Default quorum: 66%
	}
}

// AddValidator adds a new validator to the consensus system.
func (c *Consensus) AddValidator(address string, stake float64) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if _, exists := c.Validators[address]; exists {
		return errors.New("validator already exists")
	}

	c.Validators[address] = &Validator{
		Address:   address,
		Stake:     stake,
		IsTrusted: true,
		Slashable: true,
	}
	c.StakingPool += stake
	return nil
}

// RemoveValidator removes a validator from the consensus system.
func (c *Consensus) RemoveValidator(address string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	validator, exists := c.Validators[address]
	if !exists {
		return errors.New("validator not found")
	}

	c.StakingPool -= validator.Stake
	delete(c.Validators, address)
	return nil
}

// SelectValidators randomly selects a group of validators for consensus.
func (c *Consensus) SelectValidators(count int) ([]*Validator, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if count <= 0 || count > len(c.Validators) {
		return nil, errors.New("invalid number of validators to select")
	}

	validators := []*Validator{}
	for _, v := range c.Validators {
		validators = append(validators, v)
	}

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(validators), func(i, j int) {
		validators[i], validators[j] = validators[j], validators[i]
	})

	return validators[:count], nil
}

// ReachConsensus simulates reaching consensus on a proposed block.
func (c *Consensus) ReachConsensus(validators []*Validator) (bool, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if len(validators) == 0 {
		return false, errors.New("no validators provided")
	}

	yesVotes := 0
	for _, validator := range validators {
		if validator.IsTrusted {
			yesVotes++
		}
	}

	approvalRate := float64(yesVotes) / float64(len(validators))
	return approvalRate >= c.ApprovalQuorum, nil
}

// SlashValidator penalizes a validator by reducing their stake.
func (c *Consensus) SlashValidator(address string, penalty float64) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	validator, exists := c.Validators[address]
	if !exists {
		return errors.New("validator not found")
	}

	if !validator.Slashable {
		return errors.New("validator cannot be slashed")
	}

	if penalty <= 0 {
		return errors.New("penalty must be greater than zero")
	}

	if penalty > validator.Stake {
		penalty = validator.Stake
	}

	validator.Stake -= penalty
	c.StakingPool -= penalty

	if validator.Stake == 0 {
		delete(c.Validators, address)
	}

	return nil
}

// UpdateValidatorStake updates the stake of an existing validator.
func (c *Consensus) UpdateValidatorStake(address string, stake float64) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	validator, exists := c.Validators[address]
	if !exists {
		return errors.New("validator not found")
	}

	c.StakingPool -= validator.Stake
	validator.Stake = stake
	c.StakingPool += stake
	return nil
}

// GetStakingPool returns the total staking pool.
func (c *Consensus) GetStakingPool() float64 {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.StakingPool
}

// GetValidators returns the list of all validators.
func (c *Consensus) GetValidators() []*Validator {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	validators := []*Validator{}
	for _, v := range c.Validators {
		validators = append(validators, v)
	}
	return validators
}
