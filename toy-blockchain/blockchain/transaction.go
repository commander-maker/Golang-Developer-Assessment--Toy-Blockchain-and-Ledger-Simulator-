package blockchain

import "fmt"

// Transaction represents a transfer of value between two parties.
// This is the basic unit of data carried inside a Block.
type Transaction struct {
	Sender    string  `json:"sender"`
	Recipient string  `json:"recipient"`
	Amount    float64 `json:"amount"`
}

// String returns a human-readable representation of a Transaction.
func (tx Transaction) String() string {
	return fmt.Sprintf("%s → %s : %.8f", tx.Sender, tx.Recipient, tx.Amount)
}
