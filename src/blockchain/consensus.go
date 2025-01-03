package blockchain

import (
	"errors"
	"sync"
)

// VotingConsensus represents the voting-based consensus mechanism.
type VotingConsensus struct {
	Threshold float64 // Percentage of votes required to accept a block
}

// NewVotingConsensus initializes a new voting consensus mechanism.
func NewVotingConsensus(threshold float64) *VotingConsensus {
	return &VotingConsensus{Threshold: threshold}
}

// Vote represents a vote from a node.
type Vote struct {
	NodeID string // Unique ID of the node
	Approve bool  // Whether the node approves the block
}

// ProposeBlock allows a node to propose a block for voting.
func (vc *VotingConsensus) ProposeBlock(block *Block, voters []string) (bool, error) {
	var votes []Vote
	var mutex sync.Mutex
	var wg sync.WaitGroup

	// Simulate voting process
	for _, voter := range voters {
		wg.Add(1)
		go func(nodeID string) {
			defer wg.Done()
			// Simulate vote (this could involve complex logic in real cases)
			approve := vc.validateBlock(block)
			mutex.Lock()
			votes = append(votes, Vote{NodeID: nodeID, Approve: approve})
			mutex.Unlock()
		}(voter)
	}
	wg.Wait()

	// Count votes
	approvalCount := 0
	for _, vote := range votes {
		if vote.Approve {
			approvalCount++
		}
	}

	// Check if the block meets the threshold
	approvalRate := float64(approvalCount) / float64(len(voters))
	if approvalRate >= vc.Threshold {
		return true, nil
	}

	return false, errors.New("block rejected by consensus")
}

// validateBlock checks the validity of the block (basic validation for demo).
func (vc *VotingConsensus) validateBlock(block *Block) bool {
	// Add block validation logic here (e.g., hash, structure, transactions)
	return true // Simulate approval for demo purposes
}
