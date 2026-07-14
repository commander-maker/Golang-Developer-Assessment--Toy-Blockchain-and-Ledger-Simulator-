package tests

import (
	"testing"
	"toy-blockchain/blockchain"
)

func TestGenerateKeyPair(t *testing.T) {
	privateKey, err := blockchain.GenerateKeyPair()
	if err != nil {
		t.Fatalf("expected no error generating key pair, got %v", err)
	}
	if privateKey == nil {
		t.Fatal("expected non-nil private key")
	}
	if privateKey.PublicKey.X == nil || privateKey.PublicKey.Y == nil {
		t.Fatal("expected valid public key components")
	}
}

func TestSignAndVerifyTransaction(t *testing.T) {
	privateKey, err := blockchain.GenerateKeyPair()
	if err != nil {
		t.Fatalf("failed to generate key pair: %v", err)
	}

	publicKeyStr := blockchain.PublicKeyToString(&privateKey.PublicKey)

	tx := &blockchain.Transaction{
		Sender:    "Alice",
		Recipient: "Bob",
		Amount:    10.5,
		PublicKey: publicKeyStr,
	}

	signature, err := blockchain.SignTransaction(tx, privateKey)
	if err != nil {
		t.Fatalf("failed to sign transaction: %v", err)
	}
	if signature == "" {
		t.Fatal("expected non-empty signature")
	}

	tx.Signature = signature
	valid, err := blockchain.VerifyTransaction(tx)
	if err != nil {
		t.Fatalf("failed to verify transaction: %v", err)
	}
	if !valid {
		t.Fatal("expected valid signature, but verification failed")
	}
}

func TestVerifyTransactionWithModifiedData(t *testing.T) {
	privateKey, err := blockchain.GenerateKeyPair()
	if err != nil {
		t.Fatalf("failed to generate key pair: %v", err)
	}

	publicKeyStr := blockchain.PublicKeyToString(&privateKey.PublicKey)

	tx := &blockchain.Transaction{
		Sender:    "Alice",
		Recipient: "Bob",
		Amount:    10.5,
		PublicKey: publicKeyStr,
	}

	signature, err := blockchain.SignTransaction(tx, privateKey)
	if err != nil {
		t.Fatalf("failed to sign transaction: %v", err)
	}
	tx.Signature = signature

	tx.Amount = 20.0

	valid, err := blockchain.VerifyTransaction(tx)
	if err != nil {
		t.Fatalf("failed to verify transaction: %v", err)
	}
	if valid {
		t.Fatal("expected signature verification to fail after modifying transaction")
	}
}

func TestVerifyTransactionWithWrongKey(t *testing.T) {
	privateKey1, err := blockchain.GenerateKeyPair()
	if err != nil {
		t.Fatalf("failed to generate key pair 1: %v", err)
	}

	privateKey2, err := blockchain.GenerateKeyPair()
	if err != nil {
		t.Fatalf("failed to generate key pair 2: %v", err)
	}

	publicKeyStr1 := blockchain.PublicKeyToString(&privateKey1.PublicKey)

	tx := &blockchain.Transaction{
		Sender:    "Alice",
		Recipient: "Bob",
		Amount:    10.5,
		PublicKey: publicKeyStr1,
	}

	signature, err := blockchain.SignTransaction(tx, privateKey1)
	if err != nil {
		t.Fatalf("failed to sign transaction: %v", err)
	}
	tx.Signature = signature

	publicKeyStr2 := blockchain.PublicKeyToString(&privateKey2.PublicKey)
	tx.PublicKey = publicKeyStr2

	valid, err := blockchain.VerifyTransaction(tx)
	if err != nil {
		t.Fatalf("failed to verify transaction: %v", err)
	}
	if valid {
		t.Fatal("expected signature verification to fail with wrong public key")
	}
}

func TestAddTransactionWithValidSignature(t *testing.T) {
	bc := blockchain.NewBlockchain()

	privateKey, err := blockchain.GenerateKeyPair()
	if err != nil {
		t.Fatalf("failed to generate key pair: %v", err)
	}

	bc.AddTransaction(blockchain.Transaction{Sender: "SYSTEM", Recipient: "Alice", Amount: 100})

	publicKeyStr := blockchain.PublicKeyToString(&privateKey.PublicKey)

	tx := blockchain.Transaction{
		Sender:    "Alice",
		Recipient: "Bob",
		Amount:    10.5,
		PublicKey: publicKeyStr,
	}

	signature, err := blockchain.SignTransaction(&tx, privateKey)
	if err != nil {
		t.Fatalf("failed to sign transaction: %v", err)
	}
	tx.Signature = signature

	err = bc.AddTransaction(tx)
	if err != nil {
		t.Fatalf("expected transaction to be accepted, got error: %v", err)
	}

	if len(bc.PendingTransactions) != 2 {
		t.Fatalf("expected 2 pending transactions (1 SYSTEM + 1 signed), got %d", len(bc.PendingTransactions))
	}
}

func TestAddTransactionWithInvalidSignature(t *testing.T) {
	bc := blockchain.NewBlockchain()

	bc.AddTransaction(blockchain.Transaction{Sender: "SYSTEM", Recipient: "Alice", Amount: 100})

	privateKey, err := blockchain.GenerateKeyPair()
	if err != nil {
		t.Fatalf("failed to generate key pair: %v", err)
	}

	publicKeyStr := blockchain.PublicKeyToString(&privateKey.PublicKey)

	tx := blockchain.Transaction{
		Sender:    "Alice",
		Recipient: "Bob",
		Amount:    10.5,
		PublicKey: publicKeyStr,
		Signature: "invalid_signature_data",
	}

	err = bc.AddTransaction(tx)
	if err == nil {
		t.Fatal("expected transaction to be rejected due to invalid signature")
	}

	if len(bc.PendingTransactions) != 1 {
		t.Fatalf("expected 1 pending transaction (only SYSTEM), got %d", len(bc.PendingTransactions))
	}
}
