package blockchain

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
)

// Transaction represents a transfer of value between two parties.
// It now includes cryptographic proof of authorization via ECDSA signature.
type Transaction struct {
	Sender    string  `json:"sender"`
	Recipient string  `json:"recipient"`
	Amount    float64 `json:"amount"`
	PublicKey string  `json:"publicKey"`
	Signature string  `json:"signature"`
}

// String returns a human-readable representation of a Transaction.
func (tx Transaction) String() string {
	return fmt.Sprintf("%s → %s : %.8f", tx.Sender, tx.Recipient, tx.Amount)
}

// GenerateKeyPair creates a new ECDSA private/public key pair for transaction signing.
func GenerateKeyPair() (*ecdsa.PrivateKey, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}
	return privateKey, nil
}

// PublicKeyToString serializes an ECDSA public key to a hex string.
func PublicKeyToString(pub *ecdsa.PublicKey) string {
	x := pub.X.Bytes()
	y := pub.Y.Bytes()
	publicKeyBytes := append(x, y...)
	return hex.EncodeToString(publicKeyBytes)
}

// StringToPublicKey deserializes a hex string back to an ECDSA public key.
func StringToPublicKey(pubKeyStr string) (*ecdsa.PublicKey, error) {
	publicKeyBytes, err := hex.DecodeString(pubKeyStr)
	if err != nil {
		return nil, err
	}
	if len(publicKeyBytes) != 64 {
		return nil, fmt.Errorf("invalid public key length: expected 64, got %d", len(publicKeyBytes))
	}
	x := new(big.Int).SetBytes(publicKeyBytes[:32])
	y := new(big.Int).SetBytes(publicKeyBytes[32:])
	publicKey := &ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     x,
		Y:     y,
	}
	return publicKey, nil
}

// SignatureToString converts ECDSA signature (r, s) to a hex string.
func SignatureToString(r, s *big.Int) string {
	rBytes := r.Bytes()
	sBytes := s.Bytes()
	rPadded := make([]byte, 32)
	sPadded := make([]byte, 32)
	copy(rPadded[32-len(rBytes):], rBytes)
	copy(sPadded[32-len(sBytes):], sBytes)
	sigBytes := append(rPadded, sPadded...)
	return hex.EncodeToString(sigBytes)
}

// StringToSignature converts a hex string back to ECDSA signature (r, s).
func StringToSignature(sigStr string) (*big.Int, *big.Int, error) {
	sigBytes, err := hex.DecodeString(sigStr)
	if err != nil {
		return nil, nil, err
	}
	if len(sigBytes) != 64 {
		return nil, nil, fmt.Errorf("invalid signature length: expected 64, got %d", len(sigBytes))
	}
	r := new(big.Int).SetBytes(sigBytes[:32])
	s := new(big.Int).SetBytes(sigBytes[32:])
	return r, s, nil
}

// SignTransaction creates an ECDSA signature for a transaction using the private key.
// Returns the signature as a hex string.
func SignTransaction(tx *Transaction, privateKey *ecdsa.PrivateKey) (string, error) {
	messageHash := sha256.Sum256([]byte(fmt.Sprintf("%s|%s|%.8f", tx.Sender, tx.Recipient, tx.Amount)))
	r, s, err := ecdsa.Sign(rand.Reader, privateKey, messageHash[:])
	if err != nil {
		return "", err
	}
	return SignatureToString(r, s), nil
}

// VerifyTransaction checks if the transaction's signature is valid using its public key.
func VerifyTransaction(tx *Transaction) (bool, error) {
	if tx.PublicKey == "" || tx.Signature == "" {
		return false, fmt.Errorf("transaction missing public key or signature")
	}

	publicKey, err := StringToPublicKey(tx.PublicKey)
	if err != nil {
		return false, err
	}

	r, s, err := StringToSignature(tx.Signature)
	if err != nil {
		return false, err
	}

	messageHash := sha256.Sum256([]byte(fmt.Sprintf("%s|%s|%.8f", tx.Sender, tx.Recipient, tx.Amount)))
	return ecdsa.Verify(publicKey, messageHash[:], r, s), nil
}
