package tests

import (
	"testing"
	"toy-blockchain/blockchain"
)

func TestBlockchain_IsValid_CorrectChain(t *testing.T) {
	bc := blockchain.NewBlockchain()
	bc.AddBlock(singleTx("Alice", "Bob", 10.0))
	bc.AddBlock(singleTx("Bob", "Carol", 5.0))

	if !bc.IsValid() {
		t.Fatal("expected blockchain to be valid, but IsValid() returned false")
	}
}

func TestBlockchain_Validate_CorrectChain(t *testing.T) {
	bc := blockchain.NewBlockchain()
	bc.AddBlock(singleTx("Alice", "Bob", 10.0))
	bc.AddBlock(singleTx("Bob", "Carol", 5.0))

	if err := bc.Validate(); err != nil {
		t.Fatalf("expected valid blockchain, got error: %v", err)
	}
}

func TestBlockchain_Validate_InvalidHashIntegrity(t *testing.T) {
	bc := blockchain.NewBlockchain()
	bc.AddBlock(singleTx("Alice", "Bob", 10.0))
	bc.Blocks[1].Hash = "0123456789abcdef"

	if err := bc.Validate(); err == nil {
		t.Fatal("expected validation to fail for corrupted hash")
	}
}

func TestBlockchain_Validate_InvalidGenesisBlock(t *testing.T) {
	bc := blockchain.NewBlockchain()
	bc.Blocks[0].Timestamp = 123
	bc.Blocks[0].Hash = blockchain.CalculateHash(bc.Blocks[0])

	if err := bc.Validate(); err == nil {
		t.Fatal("expected validation to fail for mutated genesis block")
	}
}

func TestBlockchain_Validate_InvalidPreviousHash(t *testing.T) {
	bc := blockchain.NewBlockchain()
	bc.AddBlock(singleTx("Alice", "Bob", 10.0))
	bc.Blocks[1].PrevHash = "CORRUPTED_PREV_HASH"

	if err := bc.Validate(); err == nil {
		t.Fatal("expected validation to fail for bad previous hash")
	}
}

func TestBlockchain_Validate_InvalidProofOfWork(t *testing.T) {
	bc := blockchain.NewBlockchain()
	bc.AddBlock(singleTx("Alice", "Bob", 10.0))
	bc.Blocks[1].Hash = "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"

	if err := bc.Validate(); err == nil {
		t.Fatal("expected validation to fail for invalid proof of work")
	}
}

func TestBlockchain_Validate_InvalidBlockOrder(t *testing.T) {
	bc := blockchain.NewBlockchain()
	bc.AddBlock(singleTx("Alice", "Bob", 10.0))
	bc.Blocks[1].Index = 5
	bc.Blocks[1].Hash = blockchain.CalculateHash(bc.Blocks[1])

	if err := bc.Validate(); err == nil {
		t.Fatal("expected validation to fail for invalid block order")
	}
}

func TestBlockchain_Validate_InvalidTimestampOrder(t *testing.T) {
	bc := blockchain.NewBlockchain()
	bc.AddBlock(singleTx("Alice", "Bob", 10.0))
	bc.Blocks[1].Timestamp = bc.Blocks[0].Timestamp - 1
	bc.Blocks[1].Hash = blockchain.CalculateHash(bc.Blocks[1])

	if err := bc.Validate(); err == nil {
		t.Fatal("expected validation to fail for invalid timestamp order")
	}
}

func TestBlockchain_IsValid_TamperedData(t *testing.T) {
	bc := blockchain.NewBlockchain()
	bc.AddBlock(singleTx("Alice", "Bob", 10.0))
	bc.AddBlock(singleTx("Bob", "Carol", 5.0))

	bc.Blocks[1].Transactions[0].Amount = 999.99

	if bc.IsValid() {
		t.Fatal("expected blockchain to be invalid after modifying transaction amount, but IsValid() returned true")
	}
}

func TestBlockchain_Validate_TamperedTransactionHashMismatch(t *testing.T) {
	bc := blockchain.NewBlockchain()
	bc.AddBlock(singleTx("Alice", "Bob", 20.0))
	bc.Blocks[1].Transactions[0].Amount = 200.0

	if err := bc.Validate(); err == nil {
		t.Fatal("expected validation to fail after tampering with transaction amount, but got no error")
	}
}

func TestBlockchain_IsValid_TamperedPrevHash(t *testing.T) {
	bc := blockchain.NewBlockchain()
	bc.AddBlock(singleTx("Alice", "Bob", 10.0))
	bc.AddBlock(singleTx("Bob", "Carol", 5.0))

	bc.Blocks[2].PrevHash = "CORRUPTED_PREV_HASH"

	if bc.IsValid() {
		t.Fatal("expected blockchain to be invalid after modifying PrevHash, but IsValid() returned true")
	}
}

func TestBlockchain_IsValid_TamperedHash(t *testing.T) {
	bc := blockchain.NewBlockchain()
	bc.AddBlock(singleTx("Alice", "Bob", 10.0))

	bc.Blocks[1].Hash = "0000_CORRUPTED_HASH_THAT_MEETS_DIFFICULTY"

	if bc.IsValid() {
		t.Fatal("expected blockchain to be invalid after modifying Hash directly, but IsValid() returned true")
	}
}
