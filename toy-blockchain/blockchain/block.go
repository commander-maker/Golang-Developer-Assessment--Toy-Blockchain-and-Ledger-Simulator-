package blockchain

import (
	"fmt"
	"strings"
	"time"
)

// Block represents a single block in the blockchain.
//
// Fields (per specification):
//   - Index        : height of this block in the chain (0 = Genesis)
//   - Timestamp    : creation time as Unix nanoseconds
//   - Transactions : ordered list of transactions included in this block
//   - PrevHash     : SHA-256 hash of the preceding block (links the chain)
//   - Nonce        : arbitrary integer reserved for future Proof-of-Work mining
//   - Hash         : SHA-256 hash of this block's own contents
type Block struct {
	Index        int           `json:"index"`
	Timestamp    int64         `json:"timestamp"`
	Transactions []Transaction `json:"transactions"`
	PrevHash     string        `json:"previousHash"`
	Nonce        int           `json:"nonce"`
	Hash         string        `json:"hash"`
}

// NewBlock creates and returns a new Block.
// It automatically:
//   - sets the current Unix-nanosecond timestamp
//   - sets Nonce to 0
//   - computes the SHA-256 hash from all block fields
func NewBlock(index int, txs []Transaction, prevHash string) *Block {
	b := &Block{
		Index:        index,
		Timestamp:    time.Now().UnixNano(),
		Transactions: txs,
		PrevHash:     prevHash,
		Nonce:        0,
	}
	b.Hash = CalculateHash(b)
	return b
}

// Mine repeatedly increments the block's Nonce and calculates the hash
// until the hash satisfies the target difficulty (leading zeros).
// It also measures and prints the elapsed mining time and statistics to stdout.
func (b *Block) Mine(difficulty int) {
	fmt.Printf("Mining Block #%d...\n\n", b.Index)
	start := time.Now()
	target := strings.Repeat("0", difficulty)
	for {
		hash := CalculateHash(b)
		if strings.HasPrefix(hash, target) {
			b.Hash = hash
			break
		}
		b.Nonce++
	}
	elapsed := time.Since(start)

	fmt.Printf("Difficulty  : %d\n\n", difficulty)
	fmt.Printf("Nonce Found : %d\n\n", b.Nonce)
	fmt.Printf("Hash        :\n%s\n\n", b.Hash)
	fmt.Printf("Time        :\n%.2f seconds\n\n", elapsed.Seconds())
}
