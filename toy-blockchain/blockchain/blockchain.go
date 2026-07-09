package blockchain

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// GenesisPrevHash is the canonical previous-hash value for the Genesis block.
// It is a 64-character string of zeros — representing "no prior block".
// Using a fixed sentinel (rather than an empty string) matches the Bitcoin convention
// and makes it immediately obvious in logs/output that this is the chain origin.
const GenesisPrevHash = "0000000000000000000000000000000000000000000000000000000000000000"

// Difficulty defines the number of leading zeros required for a block's hash to be valid (Proof-of-Work).
const Difficulty = 4

// Blockchain is an ordered slice of Block pointers.
// The first element (index 0) is always the Genesis block.
type Blockchain struct {
	Blocks              []*Block      `json:"blocks"`
	PendingTransactions []Transaction `json:"pendingTransactions"`
}

// NewBlockchain is the constructor for a fresh Blockchain.
//
// It follows this exact flow, step by step:
//
//	┌─────────────────────┐
//	│   New Blockchain    │  ← allocate the Blockchain struct
//	└────────┬────────────┘
//	         ↓
//	┌─────────────────────┐
//	│ Create Genesis Block│  ← build block with fixed spec values
//	└────────┬────────────┘
//	         ↓
//	┌─────────────────────┐
//	│  Calculate Hash     │  ← SHA-256 over Index|Timestamp|Txs|PrevHash|Nonce
//	└────────┬────────────┘
//	         ↓
//	┌─────────────────────┐
//	│ Store Genesis Block │  ← append to Blocks slice
//	└────────┬────────────┘
//	         ↓
//	┌─────────────────────┐
//	│  Return Blockchain  │  ← hand back the ready chain
//	└─────────────────────┘
func NewBlockchain() *Blockchain {
	// Step 1 — New Blockchain: allocate an empty chain.
	bc := &Blockchain{
		Blocks:              []*Block{},
		PendingTransactions: []Transaction{},
	}

	// Step 2 — Create Genesis Block: build the block with all fixed spec values.
	//           genesisBlock() returns the block WITHOUT a hash yet — the hash
	//           step is deliberately kept here so the flow is explicit.
	genesis := newGenesisBlock()

	// Step 3 — Calculate Hash: compute SHA-256 over the five input fields.
	//           This is the only place the genesis hash is produced.
	genesis.Hash = CalculateHash(genesis)

	// Step 4 — Store Genesis Block: attach it as the first (and only) block.
	bc.Blocks = append(bc.Blocks, genesis)

	// Step 5 — Return Blockchain.
	return bc
}

// newGenesisBlock builds the Genesis Block struct with hard-coded spec values.
//
// It intentionally does NOT compute the hash — that is the constructor's job
// (Step 3 above). This keeps each step of the flow single-responsibility.
//
// Fixed values (per specification):
//
//	Index        = 0
//	Timestamp    = 0              (reproducible — NOT time.Now())
//	Transactions = []             (Genesis carries no transactions)
//	PrevHash     = GenesisPrevHash  (64 zeros — no prior block)
//	Nonce        = 0
func newGenesisBlock() *Block {
	return &Block{
		Index:        0,
		Timestamp:    0,
		Transactions: []Transaction{},
		PrevHash:     GenesisPrevHash,
		Nonce:        0,
		// Hash is intentionally left empty — filled in by NewBlockchain Step 3.
	}
}

// balances computes the balance for each account by walking the blockchain
// and optionally including pending transactions.
func (bc *Blockchain) balances(includePending bool) map[string]float64 {
	balances := make(map[string]float64)
	for _, block := range bc.Blocks {
		for _, tx := range block.Transactions {
			if tx.Sender != "SYSTEM" {
				balances[tx.Sender] -= tx.Amount
			}
			balances[tx.Recipient] += tx.Amount
		}
	}
	if includePending {
		for _, tx := range bc.PendingTransactions {
			if tx.Sender != "SYSTEM" {
				balances[tx.Sender] -= tx.Amount
			}
			balances[tx.Recipient] += tx.Amount
		}
	}
	return balances
}

// AddTransaction validates a transaction and adds it to the pending transactions pool.
func (bc *Blockchain) AddTransaction(tx Transaction) error {
	if tx.Amount <= 0 {
		return fmt.Errorf("transaction amount must be greater than zero, got %f", tx.Amount)
	}
	if tx.Sender != "SYSTEM" {
		balances := bc.balances(true)
		if balances[tx.Sender] < tx.Amount {
			return fmt.Errorf("insufficient balance for sender %s: has %.8f, wants to send %.8f", tx.Sender, balances[tx.Sender], tx.Amount)
		}
	}
	bc.PendingTransactions = append(bc.PendingTransactions, tx)
	return nil
}

const DefaultDataFile = "blockchain.json"

// GetBalances computes the balance for each account by walking the entire blockchain
// and accumulating the amounts from all confirmed transactions.
func (bc *Blockchain) GetBalances() map[string]float64 {
	return bc.balances(false)
}

func LoadFromFile(path string) (*Blockchain, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return NewBlockchain(), nil
		}
		return nil, err
	}

	var bc Blockchain
	if err := json.Unmarshal(data, &bc); err != nil {
		return nil, err
	}
	if len(bc.Blocks) == 0 {
		return NewBlockchain(), nil
	}
	if bc.PendingTransactions == nil {
		bc.PendingTransactions = []Transaction{}
	}
	return &bc, nil
}

func LoadBlockchain(path string) (*Blockchain, error) {
	return LoadFromFile(path)
}

func (bc *Blockchain) SaveToFile(path string) error {
	data, err := json.MarshalIndent(bc, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// AddBlock appends a new block containing the supplied transactions.
// It is a convenience wrapper for the pending-transaction flow:
//  1. Append txs to PendingTransactions
//  2. MinePendingTransactions()
func (bc *Blockchain) AddBlock(txs []Transaction) {
	bc.PendingTransactions = append(bc.PendingTransactions, txs...)
	bc.MinePendingTransactions()
}

// MinePendingTransactions creates a new block from the pending transaction pool,
// mines it using the package-level Difficulty, appends it to the chain, and then
// clears the pending pool.
func (bc *Blockchain) MinePendingTransactions() {
	if len(bc.PendingTransactions) == 0 {
		return
	}

	prev := bc.Blocks[len(bc.Blocks)-1]
	newBlock := NewBlock(prev.Index+1, bc.PendingTransactions, prev.Hash)
	newBlock.Mine(Difficulty)
	bc.Blocks = append(bc.Blocks, newBlock)
	bc.PendingTransactions = []Transaction{}
}

// Validate checks the full blockchain for integrity.
// It verifies:
//  1. Stored hash matches recalculated hash.
//  2. Each block's PrevHash matches the previous block's Hash.
//  3. Each mined block satisfies the proof-of-work difficulty.
//  4. Block indexes increase sequentially.
//  5. Block timestamps are non-decreasing.
func (bc *Blockchain) Validate() error {
	if len(bc.Blocks) == 0 {
		return fmt.Errorf("blockchain contains no blocks")
	}

	target := strings.Repeat("0", Difficulty)

	for i, current := range bc.Blocks {
		if current.Hash != CalculateHash(current) {
			return fmt.Errorf("invalid block %d: stored hash does not match calculated hash", current.Index)
		}

		if i == 0 {
			if current.Index != 0 {
				return fmt.Errorf("invalid genesis block index: expected 0, got %d", current.Index)
			}
			continue
		}

		prev := bc.Blocks[i-1]
		if current.Index != prev.Index+1 {
			return fmt.Errorf("invalid block %d index: expected %d", current.Index, prev.Index+1)
		}
		if current.PrevHash != prev.Hash {
			return fmt.Errorf("invalid block %d: previous hash does not match hash of block %d", current.Index, prev.Index)
		}
		if !strings.HasPrefix(current.Hash, target) {
			return fmt.Errorf("invalid block %d: proof of work not satisfied", current.Index)
		}
		if current.Timestamp < prev.Timestamp {
			return fmt.Errorf("invalid block %d: timestamp is earlier than previous block", current.Index)
		}
	}

	return nil
}

// IsValid walks the chain to verify that:
//  1. The link between consecutive blocks is unbroken (PrevHash == previous block's Hash).
//  2. The data of each block hasn't been tampered with (Hash matches CalculateHash).
//  3. Each mined block satisfies the difficulty requirement (except the Genesis block).
func (bc *Blockchain) IsValid() bool {
	return bc.Validate() == nil
}

// separator is the divider line used between blocks in the printed output.
const separator = "------------------------"

// Print outputs each block in the chain to stdout in the canonical format:
//
//	------------------------
//	Block #N
//	Timestamp    : <unix-nano>
//	Previous     : <64-char hex>
//	Nonce        : <int>
//	Hash         : <64-char hex>
//	Transactions : <count>
//	  [0] Alice → Bob : 10.00000000   ← only shown when count > 0
//	------------------------
//
// Use this to verify the Genesis block is created correctly (Block #0 must
// always show Timestamp=0, Previous=0000...0000, Transactions=0).
func (bc *Blockchain) Print() {
	for _, b := range bc.Blocks {
		fmt.Println(separator)
		fmt.Printf("Block #%d\n", b.Index)
		fmt.Printf("Timestamp    : %d\n", b.Timestamp)
		fmt.Printf("Previous     : %s\n", b.PrevHash)
		fmt.Printf("Nonce        : %d\n", b.Nonce)
		fmt.Printf("Hash         : %s\n", b.Hash)
		fmt.Printf("Transactions : %d\n", len(b.Transactions))

		// Print individual transactions beneath the count line.
		for i, tx := range b.Transactions {
			fmt.Printf("  [%d] %s\n", i, tx)
		}
	}
	fmt.Println(separator)
}
