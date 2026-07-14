package blockchain

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"sync"
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
	MerkleRoot   string        `json:"merkleRoot"`
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
	// Compute Merkle root from transactions before computing the block hash.
	b.MerkleRoot = CalculateMerkleRoot(b.Transactions)
	b.Hash = CalculateHash(b)
	return b
}

// Mine performs proof-of-work using multiple worker goroutines.
//
// It divides the nonce search space among several workers so that each
// worker searches unique nonces without overlap:
//
//	worker 0: 0, W, 2W, 3W, ...
//	worker 1: 1, W+1, 2W+1, 3W+1, ...
//
// where W is the worker count.
//
// The first worker to find a valid hash sends its block back over a channel,
// and cancellation is propagated to all other workers using context.
// The original block receives the successful nonce and hash before returning.
func (b *Block) Mine(difficulty int) {
	b.MineWithWorkers(difficulty, runtime.NumCPU())
}

// MineWithWorkers performs proof-of-work with configurable worker concurrency.
//
// workerCount defines how many goroutines search disjoint nonce partitions.
// A value <= 0 defaults to runtime.NumCPU().
func (b *Block) MineWithWorkers(difficulty, workerCount int) {
	b.MerkleRoot = CalculateMerkleRoot(b.Transactions)
	fmt.Printf("Mining Block #%d...\n\n", b.Index)
	start := time.Now()
	target := strings.Repeat("0", difficulty)
	if workerCount <= 0 {
		workerCount = runtime.NumCPU()
	}
	if workerCount <= 0 {
		workerCount = 1
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	resultCh := make(chan *Block, 1)
	var wg sync.WaitGroup

	// Launch workers to explore disjoint nonce partitions.
	for workerID := 0; workerID < workerCount; workerID++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			workerBlock := *b // copy block metadata, share Transactions slice safely
			workerBlock.Nonce = id

			for {
				select {
				case <-ctx.Done():
					return
				default:
				}

				hash := CalculateHash(&workerBlock)
				if strings.HasPrefix(hash, target) {
					workerBlock.Hash = hash
					select {
					case resultCh <- &workerBlock:
						cancel()
					default:
					}
					return
				}

				workerBlock.Nonce += workerCount
			}
		}(workerID)
	}

	// Wait for the first successful worker, then wait for all workers to exit.
	var winningBlock *Block
	select {
	case winningBlock = <-resultCh:
		// found a valid hash
	case <-ctx.Done():
	}
	wg.Wait()

	if winningBlock != nil {
		b.Nonce = winningBlock.Nonce
		b.Hash = winningBlock.Hash
	}

	elapsed := time.Since(start)
	fmt.Printf("Difficulty  : %d\n\n", difficulty)
	fmt.Printf("Nonce Found : %d\n\n", b.Nonce)
	fmt.Printf("Hash        :\n%s\n\n", b.Hash)
	fmt.Printf("Time        :\n%.2f seconds\n\n", elapsed.Seconds())
}
