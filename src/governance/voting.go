package governance

import (
	"errors"
	"sync"
)

// Proposal represents a proposal submitted for voting.
type Proposal struct {
	ID          string            // Unique identifier for the proposal
	Description string            // Description of the proposal
	Votes       map[string]bool   // Votes from nodes (true for yes, false for no)
	Threshold   float64           // Percentage of "yes" votes required to pass
	mutex       sync.Mutex        // Mutex for thread safety
}

// NewProposal creates a new proposal.
func NewProposal(id, description string, threshold float64) *Proposal {
	return &Proposal{
		ID:          id,
		Description: description,
		Votes:       make(map[string]bool),
		Threshold:   threshold,
	}
}

// Vote allows a node to cast a vote for the proposal.
func (p *Proposal) Vote(nodeID string, approve bool) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.Votes[nodeID] = approve
}

// CountVotes calculates the percentage of "yes" votes and determines if the proposal passes.
func (p *Proposal) CountVotes() (float64, bool) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	totalVotes := len(p.Votes)
	if totalVotes == 0 {
		return 0, false
	}

	yesVotes := 0
	for _, approve := range p.Votes {
		if approve {
			yesVotes++
		}
	}

	approvalRate := float64(yesVotes) / float64(totalVotes)
	return approvalRate, approvalRate >= p.Threshold
}
