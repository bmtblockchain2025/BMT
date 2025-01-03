package blockchain

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"math/big"
)

// Wallet represents a user's wallet containing public and private keys.
type Wallet struct {
	PrivateKey *ecdsa.PrivateKey
	PublicKey  string
	Address    string
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
	}, nil
}

// GenerateAddress creates a unique address based on the public key.
func GenerateAddress(publicKey []byte) string {
	hash := sha256.Sum256(publicKey)
	return hex.EncodeToString(hash[:])
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

// VerifySignature verifies a transaction signature using the public key.
func VerifySignature(publicKey, signature, transactionHash string) (bool, error) {
	pubKeyBytes, err := hex.DecodeString(publicKey)
	if err != nil {
		return false, err
	}

	if len(pubKeyBytes) != 64 {
		return false, errors.New("invalid public key length")
	}

	x := new(big.Int).SetBytes(pubKeyBytes[:32])
	y := new(big.Int).SetBytes(pubKeyBytes[32:])
	sigBytes, err := hex.DecodeString(signature)
	if err != nil {
		return false, err
	}

	r := new(big.Int).SetBytes(sigBytes[:len(sigBytes)/2])
	s := new(big.Int).SetBytes(sigBytes[len(sigBytes)/2:])

	hash := sha256.Sum256([]byte(transactionHash))
	isValid := ecdsa.Verify(&ecdsa.PublicKey{Curve: elliptic.P256(), X: x, Y: y}, hash[:], r, s)
	return isValid, nil
}
