package blockchain

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"math/big"
	"sync"
	"time"
)

// Contact represents a saved contact in the wallet.
type Contact struct {
	Name    string // Name of the contact (limited to 50 characters)
	Address string // Address of the contact
}

// TransactionHistory represents a detailed record of a transaction.
type TransactionHistory struct {
	Timestamp   time.Time // Time of the transaction
	To          string    // Recipient address
	Amount      float64   // Amount transferred
	Fee         float64   // Transaction fee
	Status      string    // Status of the transaction (e.g., pending, confirmed)
	IsAnonymous bool      // Whether the transaction was anonymous
}

// Wallet represents a user's wallet containing public and private keys.
type Wallet struct {
	PrivateKey    *ecdsa.PrivateKey
	PublicKey     string
	Address       string
	Balance       float64 // Balance in BMT
	Contacts      []Contact
	History       []TransactionHistory
	TransactionLimit float64
	StakedAmount  float64 // Amount of BMT staked
	mutex         sync.Mutex
}

// NewWallet creates a new wallet with a unique key pair.
func NewWallet() (*Wallet, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}

	publicKey := append(privateKey.PublicKey.X.Bytes(), privateKey.PublicKey.Y.Bytes()...)
	address := GenerateAddress(publicKey)

	return &Wallet{
		PrivateKey: privateKey,
		PublicKey:  hex.EncodeToString(publicKey),
		Address:    address,
		Balance:    0.0,
		Contacts:   []Contact{},
		History:    []TransactionHistory{},
		TransactionLimit: 1000.0, // Default transaction limit
		StakedAmount:  0.0,
	}, nil
}

// GenerateAddress creates a unique address based on the public key.
func GenerateAddress(publicKey []byte) string {
	hash := sha256.Sum256(publicKey)
	return hex.EncodeToString(hash[:])
}

// CreateAndSignTransaction creates a new transaction and signs it using the wallet's private key.
func (w *Wallet) CreateAndSignTransaction(receiver string, amount, fee float64, anonymous bool) (*Transaction, error) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if amount <= 0 {
		return nil, errors.New("amount must be greater than zero")
	}
	if amount+fee > w.Balance {
		return nil, errors.New("insufficient balance")
	}

	if err := w.CheckTransactionLimit(amount); err != nil {
		return nil, err
	}

	// Create the transaction
	tx, err := NewTransaction(w.Address, receiver, amount, fee, anonymous)
	if err != nil {
		return nil, err
	}

	// Sign the transaction
	signature, err := w.SignTransaction(tx.Hash)
	if err != nil {
		return nil, err
	}
	tx.SignTransaction(signature)

	// Add transaction to history
	w.AddTransactionHistory(receiver, amount, fee, "pending", anonymous)

	return tx, nil
}

// SignTransaction signs a transaction using the wallet's private key.
func (w *Wallet) SignTransaction(transactionHash string) (string, error) {
	hash := sha256.Sum256([]byte(transactionHash))
	r, s, err := ecdsa.Sign(rand.Reader, w.PrivateKey, hash[:])
	if err != nil {
		return "", err
	}

	signature := append(r.Bytes(), s.Bytes()...)
	return hex.EncodeToString(signature), nil
}

// StakeBMT stakes a specified amount of BMT to become a validator.
func (w *Wallet) StakeBMT(amount float64, consensus *Consensus) error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if amount <= 0 {
		return errors.New("stake amount must be greater than zero")
	}
	if amount > w.Balance {
		return errors.New("insufficient balance to stake")
	}

	w.Balance -= amount
	w.StakedAmount += amount

	return consensus.AddValidator(w.Address, w.StakedAmount)
}

// UnstakeBMT removes the staked amount and updates the balance.
func (w *Wallet) UnstakeBMT(consensus *Consensus) error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if w.StakedAmount == 0 {
		return errors.New("no staked amount to unstake")
	}

	amount := w.StakedAmount
	w.Balance += amount
	w.StakedAmount = 0

	return consensus.RemoveValidator(w.Address)
}

// AddTransactionHistory adds a transaction record to the wallet history.
func (w *Wallet) AddTransactionHistory(to string, amount, fee float64, status string, isAnonymous bool) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	w.History = append(w.History, TransactionHistory{
		Timestamp:   time.Now(),
		To:          to,
		Amount:      amount,
		Fee:         fee,
		Status:      status,
		IsAnonymous: isAnonymous,
	})
}

// GetHistory retrieves the transaction history.
func (w *Wallet) GetHistory() []TransactionHistory {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	return w.History
}

// SetTransactionLimit sets a new transaction limit.
func (w *Wallet) SetTransactionLimit(limit float64) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	w.TransactionLimit = limit
}

// CheckTransactionLimit checks if a transaction exceeds the limit.
func (w *Wallet) CheckTransactionLimit(amount float64) error {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	if amount > w.TransactionLimit {
		return errors.New("transaction exceeds the set limit")
	}
	return nil
}

// UpdateBalance updates the wallet's balance by a specified amount.
func (w *Wallet) UpdateBalance(amount float64) {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	w.Balance += amount
}

// GetBalance returns the wallet's current balance.
func (w *Wallet) GetBalance() float64 {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	return w.Balance
}
