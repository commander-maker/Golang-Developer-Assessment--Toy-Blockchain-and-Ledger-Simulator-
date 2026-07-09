package tests

import (
	"strings"
	"testing"
	"toy-blockchain/blockchain"
)

// ---------------------------------------------------------------------------
// Helper
// ---------------------------------------------------------------------------

// makeBlock builds a Block with a fixed timestamp for reproducibility.
func makeBlock(index int, txs []blockchain.Transaction, prevHash string, timestamp int64) *blockchain.Block {
	b := &blockchain.Block{
		Index:        index,
		Timestamp:    timestamp,
		Transactions: txs,
		PrevHash:     prevHash,
		Nonce:        0,
	}
	b.Hash = blockchain.CalculateHash(b)
	return b
}

// singleTx is a convenience helper for tests that need one transaction.
func singleTx(from, to string, amount float64) []blockchain.Transaction {
	return []blockchain.Transaction{{Sender: from, Recipient: to, Amount: amount}}
}

// ---------------------------------------------------------------------------
// Hash property tests
// ---------------------------------------------------------------------------

// TestCalculateHash_NotEmpty verifies CalculateHash always returns a non-empty string.
func TestCalculateHash_NotEmpty(t *testing.T) {
	b := makeBlock(0, nil, "", 1_000_000)
	if b.Hash == "" {
		t.Fatal("expected a non-empty hash, got empty string")
	}
}

// TestCalculateHash_Length verifies SHA-256 produces a 64-char hex string.
func TestCalculateHash_Length(t *testing.T) {
	b := makeBlock(0, nil, "", 1_000_000)
	if len(b.Hash) != 64 {
		t.Fatalf("expected hash length 64, got %d", len(b.Hash))
	}
}

// TestCalculateHash_Deterministic verifies the same input always yields the same hash.
func TestCalculateHash_Deterministic(t *testing.T) {
	txs := singleTx("Alice", "Bob", 10.0)
	b1 := makeBlock(1, txs, "prevABC", 9_999_999)
	b2 := makeBlock(1, txs, "prevABC", 9_999_999)

	if b1.Hash != b2.Hash {
		t.Fatalf("hashing is not deterministic:\n  got  %s\n  want %s", b2.Hash, b1.Hash)
	}
}

// TestCalculateHash_DifferentData verifies different transactions produce different hashes.
func TestCalculateHash_DifferentData(t *testing.T) {
	b1 := makeBlock(1, singleTx("Alice", "Bob", 10.0), "prevXXX", 1_234_567)
	b2 := makeBlock(1, singleTx("Alice", "Bob", 99.0), "prevXXX", 1_234_567)

	if b1.Hash == b2.Hash {
		t.Fatal("different transactions produced the same hash")
	}
}

// TestCalculateHash_DifferentPrevHash verifies changing PrevHash changes the hash.
func TestCalculateHash_DifferentPrevHash(t *testing.T) {
	txs := singleTx("Carol", "Dave", 5.0)
	b1 := makeBlock(2, txs, "hash-AAA", 5_555_555)
	b2 := makeBlock(2, txs, "hash-BBB", 5_555_555)

	if b1.Hash == b2.Hash {
		t.Fatal("different PrevHash produced the same hash — chain integrity broken")
	}
}

// TestCalculateHash_GenesisHasNoPrevHash verifies Genesis (empty PrevHash) still hashes.
func TestCalculateHash_GenesisHasNoPrevHash(t *testing.T) {
	genesis := makeBlock(0, []blockchain.Transaction{}, "", 0)
	if genesis.Hash == "" {
		t.Fatal("Genesis block should still produce a valid hash")
	}
}

// TestCalculateHash_NonceAffectsHash verifies that changing the Nonce changes the hash.
// This is critical: it proves PoW will work — incrementing Nonce changes the output.
func TestCalculateHash_NonceAffectsHash(t *testing.T) {
	txs := singleTx("Alice", "Bob", 1.0)

	b1 := &blockchain.Block{Index: 1, Timestamp: 111, Transactions: txs, PrevHash: "abc", Nonce: 0}
	b1.Hash = blockchain.CalculateHash(b1)

	b2 := &blockchain.Block{Index: 1, Timestamp: 111, Transactions: txs, PrevHash: "abc", Nonce: 1}
	b2.Hash = blockchain.CalculateHash(b2)

	if b1.Hash == b2.Hash {
		t.Fatal("different Nonce values produced the same hash — PoW would be broken")
	}
}

// TestCalculateHash_MultipleTransactions verifies that a block with multiple
// transactions hashes differently from one with a single transaction.
func TestCalculateHash_MultipleTransactions(t *testing.T) {
	oneTx := []blockchain.Transaction{{Sender: "Alice", Recipient: "Bob", Amount: 10.0}}
	twoTx := []blockchain.Transaction{
		{Sender: "Alice", Recipient: "Bob", Amount: 10.0},
		{Sender: "Bob", Recipient: "Carol", Amount: 5.0},
	}

	b1 := makeBlock(1, oneTx, "prev", 777)
	b2 := makeBlock(1, twoTx, "prev", 777)

	if b1.Hash == b2.Hash {
		t.Fatal("blocks with different transaction counts produced the same hash")
	}
}

// TestTransaction_String verifies the human-readable format of a Transaction.
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

// ---------------------------------------------------------------------------
// Constructor flow tests (Step 6)
// ---------------------------------------------------------------------------

// TestNewBlockchain_ConstructorFlow walks every step of the constructor diagram
// in order, asserting each one independently:
//
//	New Blockchain → Create Genesis → Calculate Hash → Store Genesis → Return Blockchain
func TestNewBlockchain_ConstructorFlow(t *testing.T) {
	// Step 1 — New Blockchain: calling the constructor must not panic and must
	//           return a non-nil pointer.
	bc := blockchain.NewBlockchain()
	if bc == nil {
		t.Fatal("Step 1 failed: NewBlockchain() returned nil")
	}

	// Step 2 — Create Genesis Block: the chain must contain exactly one block
	//           after construction (the Genesis block).
	if len(bc.Blocks) != 1 {
		t.Fatalf("Step 2 failed: expected 1 block after construction, got %d", len(bc.Blocks))
	}

	// Step 3 — Calculate Hash: the Genesis block must have a non-empty hash
	//           of exactly 64 hex characters (SHA-256 output).
	genesis := bc.Blocks[0]
	if genesis.Hash == "" {
		t.Fatal("Step 3 failed: Genesis block hash is empty — CalculateHash was not called")
	}
	if len(genesis.Hash) != 64 {
		t.Fatalf("Step 3 failed: Genesis hash length want 64, got %d", len(genesis.Hash))
	}

	// Step 4 — Store Genesis Block: the stored block must be index 0 with the
	//           correct spec values (proves it was stored, not just created).
	if genesis.Index != 0 {
		t.Errorf("Step 4 failed: stored block Index want 0, got %d", genesis.Index)
	}
	if genesis.PrevHash != blockchain.GenesisPrevHash {
		t.Errorf("Step 4 failed: stored block PrevHash want 64 zeros, got %s", genesis.PrevHash)
	}

	// Step 5 — Return Blockchain: the returned chain must be immediately usable.
	//           Add a block to confirm the chain is live and not read-only.
	bc.AddBlock([]blockchain.Transaction{{Sender: "Alice", Recipient: "Bob", Amount: 1.0}})
	if len(bc.Blocks) != 2 {
		t.Fatalf("Step 5 failed: chain not usable after construction — expected 2 blocks, got %d", len(bc.Blocks))
	}
}

// ---------------------------------------------------------------------------
// Genesis Block determinism tests (Step 4)
// ---------------------------------------------------------------------------

// TestGenesisBlock_IsDeterministic verifies that two separately created blockchains
// always produce the EXACT same Genesis block hash.
// This is the core Step 4 requirement: any node starting fresh must arrive at
// the identical chain origin.
func TestGenesisBlock_IsDeterministic(t *testing.T) {
	bc1 := blockchain.NewBlockchain()
	bc2 := blockchain.NewBlockchain()

	g1 := bc1.Blocks[0]
	g2 := bc2.Blocks[0]

	if g1.Hash != g2.Hash {
		t.Fatalf("Genesis hash is not deterministic:\n  chain1: %s\n  chain2: %s", g1.Hash, g2.Hash)
	}
}

// TestGenesisBlock_SpecValues verifies that the Genesis block fields exactly match
// the specification values (Index=0, Timestamp=0, PrevHash=64 zeros, Nonce=0).
func TestGenesisBlock_SpecValues(t *testing.T) {
	bc := blockchain.NewBlockchain()
	g := bc.Blocks[0]

	if g.Index != 0 {
		t.Errorf("Genesis Index: want 0, got %d", g.Index)
	}
	if g.Timestamp != 0 {
		t.Errorf("Genesis Timestamp: want 0, got %d", g.Timestamp)
	}
	if g.Nonce != 0 {
		t.Errorf("Genesis Nonce: want 0, got %d", g.Nonce)
	}
	if g.PrevHash != blockchain.GenesisPrevHash {
		t.Errorf("Genesis PrevHash: want 64 zeros, got %s", g.PrevHash)
	}
	if len(g.PrevHash) != 64 {
		t.Errorf("Genesis PrevHash length: want 64, got %d", len(g.PrevHash))
	}
	if len(g.Transactions) != 0 {
		t.Errorf("Genesis Transactions: want empty slice, got %d txs", len(g.Transactions))
	}
	if g.Hash == "" {
		t.Error("Genesis Hash must not be empty")
	}
}

// ---------------------------------------------------------------------------
// Serialization contract tests (Step 5 — most important)
// ---------------------------------------------------------------------------
// These tests prove two things:
//   1. Every field (in its specified order) independently affects the hash.
//   2. The Hash field itself is NOT part of the hash input.

// baseBlock returns a fully populated reference block used as the baseline
// for all field-order tests. Every test mutates exactly ONE field and checks
// that the resulting hash differs — proving that field is included.
func baseBlock() *blockchain.Block {
	b := &blockchain.Block{
		Index:        1,
		Timestamp:    1_000_000,
		Transactions: singleTx("Alice", "Bob", 10.0),
		PrevHash:     "abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
		Nonce:        0,
		Hash:         "", // will be computed below
	}
	b.Hash = blockchain.CalculateHash(b)
	return b
}

// TestHashInput_IndexIsIncluded proves Index is part of the hash input.
func TestHashInput_IndexIsIncluded(t *testing.T) {
	base := baseBlock()

	other := *base
	other.Index = 999
	other.Hash = blockchain.CalculateHash(&other)

	if base.Hash == other.Hash {
		t.Fatal("changing Index did not change the hash — Index is missing from hash input")
	}
}

// TestHashInput_TimestampIsIncluded proves Timestamp is part of the hash input.
func TestHashInput_TimestampIsIncluded(t *testing.T) {
	base := baseBlock()

	other := *base
	other.Timestamp = 9_999_999
	other.Hash = blockchain.CalculateHash(&other)

	if base.Hash == other.Hash {
		t.Fatal("changing Timestamp did not change the hash — Timestamp is missing from hash input")
	}
}

// TestHashInput_TransactionsIsIncluded proves Transactions are part of the hash input.
func TestHashInput_TransactionsIsIncluded(t *testing.T) {
	base := baseBlock()

	other := *base
	other.Transactions = singleTx("Alice", "Bob", 99.99) // different amount
	other.Hash = blockchain.CalculateHash(&other)

	if base.Hash == other.Hash {
		t.Fatal("changing Transactions did not change the hash — Transactions missing from hash input")
	}
}

// TestHashInput_PrevHashIsIncluded proves PreviousHash is part of the hash input.
func TestHashInput_PrevHashIsIncluded(t *testing.T) {
	base := baseBlock()

	other := *base
	other.PrevHash = "0000000000000000000000000000000000000000000000000000000000000000"
	other.Hash = blockchain.CalculateHash(&other)

	if base.Hash == other.Hash {
		t.Fatal("changing PrevHash did not change the hash — PrevHash missing from hash input")
	}
}

// TestHashInput_NonceIsIncluded proves Nonce is part of the hash input.
func TestHashInput_NonceIsIncluded(t *testing.T) {
	base := baseBlock()

	other := *base
	other.Nonce = 42
	other.Hash = blockchain.CalculateHash(&other)

	if base.Hash == other.Hash {
		t.Fatal("changing Nonce did not change the hash — Nonce missing from hash input")
	}
}

// TestHashInput_HashFieldIsExcluded proves the Hash field is NOT fed back
// into CalculateHash. If it were, calling CalculateHash twice on the same
// block would produce two different results (circular dependency).
func TestHashInput_HashFieldIsExcluded(t *testing.T) {
	// Build a block and compute its hash once.
	b := baseBlock()
	firstHash := b.Hash

	// Artificially corrupt b.Hash to a random sentinel value.
	// If CalculateHash reads b.Hash as part of its input, the next call
	// will return a different result.
	b.Hash = "THIS_SHOULD_NOT_AFFECT_THE_OUTPUT"
	secondHash := blockchain.CalculateHash(b)

	if firstHash != secondHash {
		t.Fatalf(
			"Hash field appears to be included in hash input (circular!):\n  before: %s\n  after:  %s",
			firstHash, secondHash,
		)
	}
}

// TestHashInput_FieldOrderMatters proves that field order is stable and significant.
// We construct the raw canonical string manually using the documented order
// Index→Timestamp→Transactions→PreviousHash→Nonce and verify the hash matches.
func TestHashInput_FieldOrderMatters(t *testing.T) {
	txs := singleTx("Alice", "Bob", 10.0)

	// Block A: Index=1, Index-heavy fields first
	a := &blockchain.Block{Index: 1, Timestamp: 2, Transactions: txs, PrevHash: "ph", Nonce: 3}
	a.Hash = blockchain.CalculateHash(a)

	// Block B: same content but Index and Nonce swapped — different block entirely
	b := &blockchain.Block{Index: 3, Timestamp: 2, Transactions: txs, PrevHash: "ph", Nonce: 1}
	b.Hash = blockchain.CalculateHash(b)

	if a.Hash == b.Hash {
		t.Fatal("swapping field values produced the same hash — field order is not enforced")
	}
}

// TestBlock_Mine verifies that mining a block produces a hash satisfying the difficulty.
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

	// Verify that the hash actually matches the fields in the block (including the found nonce)
	recalculatedHash := blockchain.CalculateHash(b)
	if b.Hash != recalculatedHash {
		t.Fatalf("mined block hash %q does not match recalculated hash %q", b.Hash, recalculatedHash)
	}
}

// TestBlockchain_IsValid_CorrectChain verifies that a freshly created blockchain
// with added blocks is considered valid.
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

// TestBlockchain_IsValid_TamperedData verifies that changing transaction data
// in any block causes the blockchain to be invalid.
func TestBlockchain_IsValid_TamperedData(t *testing.T) {
	bc := blockchain.NewBlockchain()
	bc.AddBlock(singleTx("Alice", "Bob", 10.0))
	bc.AddBlock(singleTx("Bob", "Carol", 5.0))

	// Tamper with the transaction amount of block 1
	bc.Blocks[1].Transactions[0].Amount = 999.99

	if bc.IsValid() {
		t.Fatal("expected blockchain to be invalid after modifying transaction amount, but IsValid() returned true")
	}
}

func TestBlockchain_Validate_TamperedTransactionHashMismatch(t *testing.T) {
	bc := blockchain.NewBlockchain()
	bc.AddBlock(singleTx("Alice", "Bob", 20.0))

	// Tamper with the transaction amount in block 1, without updating the block hash.
	bc.Blocks[1].Transactions[0].Amount = 200.0

	if err := bc.Validate(); err == nil {
		t.Fatal("expected validation to fail after tampering with transaction amount, but got no error")
	}
}

// TestBlockchain_IsValid_TamperedPrevHash verifies that changing PrevHash in a block
// (breaking the cryptographic linkage) causes the blockchain to be invalid.
func TestBlockchain_IsValid_TamperedPrevHash(t *testing.T) {
	bc := blockchain.NewBlockchain()
	bc.AddBlock(singleTx("Alice", "Bob", 10.0))
	bc.AddBlock(singleTx("Bob", "Carol", 5.0))

	// Break the link of block 2 pointing to block 1
	bc.Blocks[2].PrevHash = "CORRUPTED_PREV_HASH"

	if bc.IsValid() {
		t.Fatal("expected blockchain to be invalid after modifying PrevHash, but IsValid() returned true")
	}
}

// TestBlockchain_IsValid_TamperedHash verifies that modifying a block's Hash directly
// causes the blockchain to be invalid.
func TestBlockchain_IsValid_TamperedHash(t *testing.T) {
	bc := blockchain.NewBlockchain()
	bc.AddBlock(singleTx("Alice", "Bob", 10.0))

	// Modify the hash of block 1
	bc.Blocks[1].Hash = "0000_CORRUPTED_HASH_THAT_MEETS_DIFFICULTY"

	if bc.IsValid() {
		t.Fatal("expected blockchain to be invalid after modifying Hash directly, but IsValid() returned true")
	}
}
