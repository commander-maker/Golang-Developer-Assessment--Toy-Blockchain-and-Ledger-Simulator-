package tests

import (
	"strings"
	"testing"
	"toy-blockchain/blockchain"
)

func TestTransaction_String(t *testing.T) {
	tx := blockchain.Transaction{Sender: "Alice", Recipient: "Bob", Amount: 10.5}
	s := tx.String()

	if !strings.Contains(s, "Alice") || !strings.Contains(s, "Bob") {
		t.Fatalf("Transaction.String() missing expected names: %s", s)
	}
	if !strings.Contains(s, "10.50000000") {
		t.Fatalf("Transaction.String() missing amount: %s", s)
	}
}

func TestAddTransaction_RejectsNegativeAmount(t *testing.T) {
	bc := blockchain.NewBlockchain()
	err := bc.AddTransaction(blockchain.Transaction{Sender: "Alice", Recipient: "Bob", Amount: -5})
	if err == nil {
		t.Fatal("expected error for negative transaction amount")
	}
}

func TestAddTransaction_RejectsZeroAmount(t *testing.T) {
	bc := blockchain.NewBlockchain()
	err := bc.AddTransaction(blockchain.Transaction{Sender: "Alice", Recipient: "Bob", Amount: 0})
	if err == nil {
		t.Fatal("expected error for zero transaction amount")
	}
}

func TestAddTransaction_RejectsInsufficientBalance(t *testing.T) {
	bc := blockchain.NewBlockchain()
	err := bc.AddTransaction(blockchain.Transaction{Sender: "Alice", Recipient: "Bob", Amount: 100})
	if err == nil {
		t.Fatal("expected error for insufficient balance")
	}
}

func TestAddTransaction_AllowsPendingFundsToBeSpent(t *testing.T) {
	bc := blockchain.NewBlockchain()

	if err := bc.AddTransaction(blockchain.Transaction{Sender: "SYSTEM", Recipient: "Alice", Amount: 20}); err != nil {
		t.Fatal(err)
	}
	if err := bc.AddTransaction(blockchain.Transaction{Sender: "Alice", Recipient: "Bob", Amount: 10}); err != nil {
		t.Fatal(err)
	}

	if err := bc.AddTransaction(blockchain.Transaction{Sender: "Alice", Recipient: "Carol", Amount: 15}); err == nil {
		t.Fatal("expected insufficient balance error for Alice after pending spend")
	}
}
