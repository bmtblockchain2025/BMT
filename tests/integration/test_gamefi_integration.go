package integration_test

import (
	"BMT-Blockchain/src/applications/gamefi"
	"testing"
)

func TestGameFiIntegration(t *testing.T) {
	// Initialize NFT Marketplace
	marketplace := gamefi.NewNFTMarketplace()

	// Mint NFT
	nft, err := marketplace.MintNFT("Alice", "NFT001", "Sword", "A legendary sword")
	if err != nil {
		t.Errorf("Failed to mint NFT: %v", err)
	}

	// List NFT for Sale
	err = marketplace.ListNFT("Alice", "NFT001", 100.0)
	if err != nil {
		t.Errorf("Failed to list NFT for sale: %v", err)
	}

	// Buy NFT
	err = marketplace.BuyNFT("Bob", "NFT001", 100.0)
	if err != nil {
		t.Errorf("Failed to buy NFT: %v", err)
	}

	// Check Ownership
	owner, err := marketplace.GetNFTOwner("NFT001")
	if err != nil {
		t.Errorf("Failed to get NFT owner: %v", err)
	}
	if owner != "Bob" {
		t.Errorf("Expected NFT owner to be Bob, got %s", owner)
	}
}
