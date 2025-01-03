package gamefi

import (
	"errors"
	"sync"
)

// Listing represents an NFT listed for sale in the marketplace.
type Listing struct {
	NFTID  string  // NFT ID being sold
	Seller string  // Owner of the NFT
	Price  float64 // Price in BMT
}

// NFTMarketplace manages listings and transactions for NFTs.
type NFTMarketplace struct {
	Listings map[string]*Listing // Mapping from NFT ID to listing
	NFTs     map[string]*NFT     // NFT storage
	mutex    sync.Mutex          // Mutex for thread safety
}

// NewNFTMarketplace initializes a new NFT marketplace.
func NewNFTMarketplace() *NFTMarketplace {
	return &NFTMarketplace{
		Listings: make(map[string]*Listing),
		NFTs:     make(map[string]*NFT),
	}
}

// AddNFT adds an NFT to the marketplace (e.g., after minting).
func (m *NFTMarketplace) AddNFT(nft *NFT) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.NFTs[nft.ID] = nft
}

// ListNFT lists an NFT for sale in the marketplace.
func (m *NFTMarketplace) ListNFT(seller, nftID string, price float64) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	nft, exists := m.NFTs[nftID]
	if !exists {
		return errors.New("NFT not found")
	}
	if nft.Owner != seller {
		return errors.New("only the owner can list the NFT for sale")
	}
	if _, listed := m.Listings[nftID]; listed {
		return errors.New("NFT is already listed for sale")
	}

	m.Listings[nftID] = &Listing{
		NFTID:  nftID,
		Seller: seller,
		Price:  price,
	}
	return nil
}

// BuyNFT allows a buyer to purchase an NFT from the marketplace.
func (m *NFTMarketplace) BuyNFT(buyer, nftID string, payment float64) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	listing, exists := m.Listings[nftID]
	if !exists {
		return errors.New("NFT is not listed for sale")
	}
	if payment < listing.Price {
		return errors.New("insufficient payment")
	}

	nft, exists := m.NFTs[nftID]
	if !exists {
		return errors.New("NFT not found")
	}

	// Transfer ownership
	nft.Owner = buyer

	// Remove the listing
	delete(m.Listings, nftID)
	return nil
}

// RemoveListing removes an NFT listing from the marketplace.
func (m *NFTMarketplace) RemoveListing(seller, nftID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	listing, exists := m.Listings[nftID]
	if !exists {
		return errors.New("NFT is not listed for sale")
	}
	if listing.Seller != seller {
		return errors.New("only the seller can remove the listing")
	}

	delete(m.Listings, nftID)
	return nil
}

// GetListing retrieves a listing by NFT ID.
func (m *NFTMarketplace) GetListing(nftID string) (*Listing, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	listing, exists := m.Listings[nftID]
	if !exists {
		return nil, errors.New("listing not found")
	}

	return listing, nil
}
