package tests

import (
	"strings"
	"testing"
	"toy-blockchain/blockchain"
)

func TestMinePendingTransactions_CreatesBlockFromPendingPool(t *testing.T) {
	bc := blockchain.NewBlockchain()

	if err := bc.AddTransaction(blockchain.Transaction{Sender: "SYSTEM", Recipient: "Alice", Amount: 50}); err != nil {
		t.Fatal(err)
	}
	if err := bc.AddTransaction(blockchain.Transaction{Sender: "Alice", Recipient: "Bob", Amount: 20}); err != nil {
		t.Fatal(err)
	}
	if err := bc.AddTransaction(blockchain.Transaction{Sender: "Alice", Recipient: "Charlie", Amount: 10}); err != nil {
		t.Fatal(err)
	}

	bc.MinePendingTransactions()

	if len(bc.Blocks) != 2 {
		t.Fatalf("expected 2 blocks after mining pending transactions, got %d", len(bc.Blocks))
	}

	minedBlock := bc.Blocks[1]
	if len(minedBlock.Transactions) != 3 {
		t.Fatalf("expected mined block to contain 3 transactions, got %d", len(minedBlock.Transactions))
	}

	if len(bc.PendingTransactions) != 0 {
		t.Fatalf("expected pending pool to be cleared after mining, got %d pending transactions", len(bc.PendingTransactions))
	}
}

func TestBlock_Mine(t *testing.T) {
	txs := singleTx("Alice", "Bob", 10.0)
	b := &blockchain.Block{
		Index:        1,
		Timestamp:    1234567,
		Transactions: txs,
		PrevHash:     "prevHash",
		Nonce:        0,
	}

	difficulty := 3
	b.Mine(difficulty)

	expectedPrefix := strings.Repeat("0", difficulty)
	if !strings.HasPrefix(b.Hash, expectedPrefix) {
		t.Fatalf("expected mined block hash to start with %q, got %q", expectedPrefix, b.Hash)
	}

	recalculatedHash := blockchain.CalculateHash(b)
	if b.Hash != recalculatedHash {
		t.Fatalf("mined block hash %q does not match recalculated hash %q", b.Hash, recalculatedHash)
	}
}
