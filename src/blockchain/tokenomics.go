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
	TransactionMutex sync.Mutex         // Mutex for thread-safe operations
}

// NewTokenomics initializes the tokenomics with total supply and max supply.
func NewTokenomics(totalSupply, maxSupply float64) *Tokenomics {
	return &Tokenomics{
		TotalSupply: totalSupply,
		MaxSupply:   maxSupply,
		Balances:    make(map[string]float64),
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

// GetBalance retrieves the balance of a specific wallet.
func (t *Tokenomics) GetBalance(address string) float64 {
	t.TransactionMutex.Lock()
	defer t.TransactionMutex.Unlock()
	return t.Balances[address]
}
