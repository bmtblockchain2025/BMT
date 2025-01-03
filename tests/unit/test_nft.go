package gamefi_test

import (
	"BMT-Blockchain/src/applications/gamefi"
	"testing"
)

func TestMintAndTransferNFT(t *testing.T) {
	marketplace := gamefi.NewNFTMarketplace()

	// Mint NFT
	nft, err := marketplace.MintNFT("Alice", "NFT001", "Sword", "A powerful sword")
	if err != nil {
		t.Errorf("Failed to mint NFT: %v", err)
	}
	if nft.Owner != "Alice" {
		t.Errorf("Expected NFT owner to be Alice, got %s", nft.Owner)
	}

	// Transfer NFT
	err = marketplace.TransferNFT("NFT001", "Bob")
	if err != nil {
		t.Errorf("Failed to transfer NFT: %v", err)
	}
	owner, err := marketplace.GetNFTOwner("NFT001")
	if err != nil || owner != "Bob" {
		t.Errorf("Expected NFT owner to be Bob, got %s", owner)
	}
}
