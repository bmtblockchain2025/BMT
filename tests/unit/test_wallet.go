package blockchain_test

import (
	"BMT-Blockchain/src/blockchain"
	"testing"
)

func TestNewWallet(t *testing.T) {
	wallet, err := blockchain.NewWallet()
	if err != nil {
		t.Errorf("Failed to create wallet: %v", err)
	}
	if wallet.Address == "" {
		t.Error("Wallet address should not be empty")
	}
	if wallet.PublicKey == "" {
		t.Error("Wallet public key should not be empty")
	}
}

func TestSignAndVerifyTransaction(t *testing.T) {
	wallet, _ := blockchain.NewWallet()
	transactionHash := "sample_transaction_hash"

	// Sign transaction
	signature, err := wallet.SignTransaction(transactionHash)
	if err != nil {
		t.Errorf("Failed to sign transaction: %v", err)
	}

	// Verify signature
	isValid, err := blockchain.VerifySignature(wallet.PublicKey, signature, transactionHash)
	if err != nil {
		t.Errorf("Failed to verify signature: %v", err)
	}
	if !isValid {
		t.Error("Signature verification failed")
	}
}
