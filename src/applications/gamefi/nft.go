package gamefi

import (
	"errors"
	"sync"
)

// NFT represents a single NFT with metadata.
type NFT struct {
	ID          string // Unique identifier for the NFT
	Owner       string // Current owner of the NFT
	Name        string // Name of the NFT
	Description string // Description of the NFT
}

// NFTMarketplace manages NFTs and their transactions.
type NFTMarketplace struct {
	NFTs   map[string]*NFT // Mapping from NFT ID to NFT object
	mutex  sync.Mutex      // Mutex for thread safety
}

// NewNFTMarketplace initializes a new NFT marketplace.
func NewNFTMarketplace() *NFTMarketplace {
	return &NFTMarketplace{
		NFTs: make(map[string]*NFT),
	}
}

// MintNFT creates a new NFT and assigns it to the specified owner.
func (m *NFTMarketplace) MintNFT(owner, id, name, description string) (*NFT, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, exists := m.NFTs[id]; exists {
		return nil, errors.New("NFT with this ID already exists")
	}

	nft := &NFT{
		ID:          id,
		Owner:       owner,
		Name:        name,
		Description: description,
	}
	m.NFTs[id] = nft
	return nft, nil
}

// TransferNFT transfers ownership of an NFT to a new owner.
func (m *NFTMarketplace) TransferNFT(nftID, newOwner string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	nft, exists := m.NFTs[nftID]
	if !exists {
		return errors.New("NFT not found")
	}

	if nft.Owner == newOwner {
		return errors.New("new owner is already the current owner")
	}

	nft.Owner = newOwner
	return nil
}

// GetNFTOwner retrieves the current owner of an NFT.
func (m *NFTMarketplace) GetNFTOwner(nftID string) (string, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	nft, exists := m.NFTs[nftID]
	if !exists {
		return "", errors.New("NFT not found")
	}

	return nft.Owner, nil
}

// GetNFT retrieves an NFT by its ID.
func (m *NFTMarketplace) GetNFT(nftID string) (*NFT, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	nft, exists := m.NFTs[nftID]
	if !exists {
		return nil, errors.New("NFT not found")
	}

	return nft, nil
}
