package blockchain

import (
	"crypto/sha256"
	"fmt"
	"strings"
)

// CalculateHash produces a deterministic SHA-256 hash for a given Block.
//
// # Serialization contract
//
// Fields are joined in this exact order with "|" as a separator:
//
//	Index | Timestamp | Transactions | PreviousHash | Nonce
//
// The Hash field itself is intentionally EXCLUDED — it is the output of
// this function, not part of its input. Including it would be circular.
//
// Each field is formatted as follows:
//
//	Index        → decimal integer          e.g. "3"
//	Timestamp    → decimal integer          e.g. "1783319207065260400"
//	Transactions → pipe-delimited records,  e.g. "Alice|Bob|10.00000000|Carol|Dave|5.00000000"
//	               each tx: Sender|Recipient|Amount(8dp)
//	PreviousHash → hex string as-is         e.g. "0000...0000"
//	Nonce        → decimal integer          e.g. "0"
//
// The "|" separator between every field guarantees no two distinct inputs
// can ever produce the same byte sequence (prevents length-extension collisions).
func CalculateHash(b *Block) string {
	record := strings.Join([]string{
		fmt.Sprintf("%d", b.Index),
		fmt.Sprintf("%d", b.Timestamp),
		serializeTransactions(b.Transactions),
		b.PrevHash,
		fmt.Sprintf("%d", b.Nonce),
	}, "|")

	hash := sha256.Sum256([]byte(record))
	return fmt.Sprintf("%x", hash) // 64 lowercase hex chars
}

// serializeTransactions converts a slice of Transactions into a single
// deterministic string for inclusion in the hash input.
//
// Format: each transaction is encoded as "Sender|Recipient|Amount" (amount to 8 d.p.)
// and all transactions are joined with "|". An empty slice → "".
//
// Example (2 transactions):
//
//	"Alice|Bob|10.00000000|Carol|Dave|5.50000000"
func serializeTransactions(txs []Transaction) string {
	if len(txs) == 0 {
		return ""
	}
	var sb strings.Builder
	for i, tx := range txs {
		if i > 0 {
			sb.WriteString("|")
		}
		sb.WriteString(fmt.Sprintf("%s|%s|%.8f", tx.Sender, tx.Recipient, tx.Amount))
	}
	return sb.String()
}
