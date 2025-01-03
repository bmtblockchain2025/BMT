package governance_test

import (
	"BMT-Blockchain/src/governance"
	"testing"
)

func TestProposalCreation(t *testing.T) {
	proposal := governance.NewProposal("P1", "Increase block size", 0.66)
	if proposal.ID != "P1" {
		t.Errorf("Expected proposal ID 'P1', got %s", proposal.ID)
	}
	if proposal.Threshold != 0.66 {
		t.Errorf("Expected threshold 0.66, got %f", proposal.Threshold)
	}
}

func TestVoting(t *testing.T) {
	proposal := governance.NewProposal("P1", "Enable staking", 0.75)

	// Cast votes
	proposal.Vote("Node1", true)
	proposal.Vote("Node2", false)
	proposal.Vote("Node3", true)

	// Count votes
	approvalRate, passed := proposal.CountVotes()
	if approvalRate < 0.66 {
		t.Errorf("Expected approval rate >= 0.66, got %f", approvalRate)
	}
	if !passed {
		t.Error("Proposal should have passed with 66% threshold")
	}
}
