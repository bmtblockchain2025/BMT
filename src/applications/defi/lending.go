package defi

import (
	"errors"
	"sync"
)

// LendingPool represents a pool for lending and borrowing assets.
type LendingPool struct {
	TotalSupply int64              // Total supply of assets in the pool
	TotalBorrowed int64            // Total amount borrowed
	InterestRate float64           // Annual interest rate
	Balances     map[string]int64  // User balances (deposited assets)
	Borrowed     map[string]int64  // User borrow amounts
	mutex        sync.Mutex        // Mutex for thread safety
}

// NewLendingPool creates a new lending pool with initial supply and interest rate.
func NewLendingPool(initialSupply int64, interestRate float64) *LendingPool {
	return &LendingPool{
		TotalSupply:  initialSupply,
		InterestRate: interestRate,
		Balances:     make(map[string]int64),
		Borrowed:     make(map[string]int64),
	}
}

// Deposit adds assets to the pool and updates the user's balance.
func (lp *LendingPool) Deposit(user string, amount int64) error {
	lp.mutex.Lock()
	defer lp.mutex.Unlock()

	if amount <= 0 {
		return errors.New("amount must be greater than zero")
	}

	lp.TotalSupply += amount
	lp.Balances[user] += amount
	return nil
}

// Borrow allows a user to borrow assets from the pool if sufficient collateral exists.
func (lp *LendingPool) Borrow(user string, amount int64) error {
	lp.mutex.Lock()
	defer lp.mutex.Unlock()

	if amount <= 0 {
		return errors.New("amount must be greater than zero")
	}
	if amount > lp.TotalSupply {
		return errors.New("insufficient liquidity in the pool")
	}

	// Simple rule: user can borrow up to 50% of their deposited balance.
	collateral := lp.Balances[user]
	if amount > collateral/2 {
		return errors.New("insufficient collateral")
	}

	lp.TotalSupply -= amount
	lp.TotalBorrowed += amount
	lp.Borrowed[user] += amount
	return nil
}

// Repay allows a user to repay borrowed assets.
func (lp *LendingPool) Repay(user string, amount int64) error {
	lp.mutex.Lock()
	defer lp.mutex.Unlock()

	if amount <= 0 {
		return errors.New("amount must be greater than zero")
	}
	if amount > lp.Borrowed[user] {
		return errors.New("repayment exceeds borrowed amount")
	}

	lp.TotalSupply += amount
	lp.TotalBorrowed -= amount
	lp.Borrowed[user] -= amount
	return nil
}

// GetUserBalance retrieves the user's deposited balance.
func (lp *LendingPool) GetUserBalance(user string) int64 {
	lp.mutex.Lock()
	defer lp.mutex.Unlock()
	return lp.Balances[user]
}

// GetUserBorrowed retrieves the user's borrowed amount.
func (lp *LendingPool) GetUserBorrowed(user string) int64 {
	lp.mutex.Lock()
	defer lp.mutex.Unlock()
	return lp.Borrowed[user]
}
