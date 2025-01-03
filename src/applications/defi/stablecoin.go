package defi

import (
	"errors"
	"sync"
)

// Stablecoin represents a stablecoin management system.
type Stablecoin struct {
	TotalSupply float64          // Total supply of the stablecoin
	Reserves    float64          // Total reserves backing the stablecoin
	PegRate     float64          // Peg rate (e.g., 1 stablecoin = 1 USD)
	Balances    map[string]float64 // User balances
	mutex       sync.Mutex       // Mutex for thread safety
}

// NewStablecoin initializes a new stablecoin system.
func NewStablecoin(initialReserves float64, pegRate float64) *Stablecoin {
	return &Stablecoin{
		Reserves: initialReserves,
		PegRate:  pegRate,
		Balances: make(map[string]float64),
	}
}

// Mint allows users to mint new stablecoins by depositing reserves.
func (sc *Stablecoin) Mint(user string, amount float64) error {
	sc.mutex.Lock()
	defer sc.mutex.Unlock()

	if amount <= 0 {
		return errors.New("mint amount must be greater than zero")
	}
	if amount > sc.Reserves {
		return errors.New("insufficient reserves to mint stablecoins")
	}

	sc.Reserves -= amount
	mintedStablecoins := amount / sc.PegRate
	sc.TotalSupply += mintedStablecoins
	sc.Balances[user] += mintedStablecoins
	return nil
}

// Burn allows users to burn stablecoins and reclaim reserves.
func (sc *Stablecoin) Burn(user string, amount float64) (float64, error) {
	sc.mutex.Lock()
	defer sc.mutex.Unlock()

	if amount <= 0 {
		return 0, errors.New("burn amount must be greater than zero")
	}
	if amount > sc.Balances[user] {
		return 0, errors.New("burn amount exceeds user balance")
	}

	reclaimedReserves := amount * sc.PegRate
	sc.Reserves += reclaimedReserves
	sc.TotalSupply -= amount
	sc.Balances[user] -= amount
	return reclaimedReserves, nil
}

// GetBalance retrieves the balance of a user.
func (sc *Stablecoin) GetBalance(user string) float64 {
	sc.mutex.Lock()
	defer sc.mutex.Unlock()
	return sc.Balances[user]
}

// GetTotalSupply retrieves the total supply of the stablecoin.
func (sc *Stablecoin) GetTotalSupply() float64 {
	sc.mutex.Lock()
	defer sc.mutex.Unlock()
	return sc.TotalSupply
}

// GetReserves retrieves the total reserves backing the stablecoin.
func (sc *Stablecoin) GetReserves() float64 {
	sc.mutex.Lock()
	defer sc.mutex.Unlock()
	return sc.Reserves
}
