# ECDSA Signature-Based Transaction Authentication

## Overview

The blockchain now implements production-style digital signature authentication using ECDSA (Elliptic Curve Digital Signature Algorithm). This ensures that:

- **Only the transaction sender** can authorize transfers from their account
- **Transactions cannot be tampered with** after signing
- **Recipients cannot forge transactions** claiming to be from other users

## How It Works

### 1. Key Generation

Each user has a unique ECDSA key pair:

```go
privateKey, err := blockchain.GenerateKeyPair()
publicKeyStr := blockchain.PublicKeyToString(&privateKey.PublicKey)
```

- **Private Key**: Kept secret by the user; used to sign transactions
- **Public Key**: Shared with the network; used to verify signatures

### 2. Transaction Signing

Before sending a transaction, the sender signs it with their private key:

```go
tx := blockchain.Transaction{
    Sender:    "Alice",
    Recipient: "Bob",
    Amount:    50.0,
    PublicKey: alicePublicKeyStr,  // Recipient needs this to verify
}

signature, err := blockchain.SignTransaction(&tx, alicePrivateKey)
tx.Signature = signature
```

The signature proves:
- Alice authorized this specific transaction
- The amount, recipient, and sender cannot be changed

### 3. Transaction Verification

When a transaction is added to the blockchain, the signature is automatically verified:

```go
blockchain.AddTransaction(tx)  // Internally calls VerifyTransaction(&tx)
```

Verification checks:
- The signature is cryptographically valid
- It matches the transaction data (sender, recipient, amount)
- It was created with the corresponding private key

### 4. Special Case: SYSTEM Transactions

Mining rewards from "SYSTEM" don't require signatures because:
- They're not user-authorized (created by the protocol)
- The blockchain trusts its own reward mechanism

## Transaction Structure

```go
type Transaction struct {
    Sender    string  `json:"sender"`
    Recipient string  `json:"recipient"`
    Amount    float64 `json:"amount"`
    PublicKey string  `json:"publicKey"`    // NEW: Public key for verification
    Signature string  `json:"signature"`    // NEW: ECDSA signature (hex-encoded)
}
```

## Security Properties

### Valid Transaction Example
```
Alice's Private Key + Transaction Data → Valid Signature
                                             ↓
                                       VerifyTransaction()
                                             ↓
                                        ✓ ACCEPTED
```

### Tampered Transaction Example
```
Modified Amount (50 → 100)
                ↓
Alice's Signature still validates 50
                ↓
                ✗ REJECTED (signature doesn't match data)
```

### Wrong Key Example
```
Bob's Public Key used to verify Alice's Signature
                ↓
                ✗ REJECTED (doesn't match Alice's signature)
```

## API Reference

### GenerateKeyPair
```go
privateKey, err := blockchain.GenerateKeyPair()
```
Creates a new ECDSA key pair. Returns nil error on success.

### PublicKeyToString / StringToPublicKey
```go
pubStr := blockchain.PublicKeyToString(&privateKey.PublicKey)
pubKey, err := blockchain.StringToPublicKey(pubStr)
```
Serialize/deserialize public keys to/from hex format for JSON storage.

### SignTransaction
```go
signature, err := blockchain.SignTransaction(&tx, privateKey)
```
Signs a transaction. Returns hex-encoded signature string.

### VerifyTransaction
```go
valid, err := blockchain.VerifyTransaction(&tx)
```
Verifies transaction signature. Returns true if valid, false otherwise.

### AddTransaction
```go
err := blockchain.AddTransaction(tx)
```
Automatically verifies signature for non-SYSTEM transactions. Rejects if:
- Amount ≤ 0
- Signature is invalid
- Sender has insufficient balance

## Error Handling

Common errors:

| Error | Cause | Solution |
|-------|-------|----------|
| `transaction missing public key or signature` | User forgot to sign | Call `SignTransaction()` |
| `invalid transaction signature` | Data was modified after signing | Resign the transaction |
| `signature verification error: ...` | Key format invalid | Ensure keys from same `GenerateKeyPair()` call |
| `insufficient balance` | Not enough funds (including pending) | Request more funds first |

## Example: Full Transaction Flow

```go
// 1. Setup
bc := blockchain.NewBlockchain()
alicePrivateKey, _ := blockchain.GenerateKeyPair()
alicePublicKeyStr := blockchain.PublicKeyToString(&alicePrivateKey.PublicKey)

// 2. Alice receives initial funds (SYSTEM - no signature needed)
bc.AddTransaction(blockchain.Transaction{
    Sender:    "SYSTEM",
    Recipient: "Alice",
    Amount:    100,
})

// 3. Alice sends to Bob
tx := blockchain.Transaction{
    Sender:    "Alice",
    Recipient: "Bob",
    Amount:    50,
    PublicKey: alicePublicKeyStr,
}
signature, _ := blockchain.SignTransaction(&tx, alicePrivateKey)
tx.Signature = signature

// 4. Add to blockchain (signature automatically verified)
bc.AddTransaction(tx)  // ✓ Success

// 5. Attempt tampering (signature becomes invalid)
tx.Amount = 200  // Changed after signing
bc.AddTransaction(tx)  // ✗ Fails: invalid signature
```

## Testing

Comprehensive tests verify:
- Valid signatures are accepted
- Modified transactions are rejected
- Wrong keys are rejected
- Signature generation is deterministic
- Integration with AddTransaction validation chain

Run tests:
```bash
go test ./tests -v
# All signature tests pass: ✓
```

## Cryptographic Details

- **Algorithm**: ECDSA (Elliptic Curve Digital Signature Algorithm)
- **Curve**: P-256 (NIST standard)
- **Hash**: SHA-256 (applied to transaction data before signing)
- **Encoding**: Hex for all keys and signatures
- **Public Key Size**: 64 bytes (32-byte X + 32-byte Y coordinates)
- **Signature Size**: 64 bytes (32-byte r + 32-byte s components)

## Production Considerations

This implementation is suitable for:
- ✓ Educational demonstrations
- ✓ Testing concepts
- ✓ Prototyping

For production systems, also implement:
- Secure key storage (not in memory)
- Key rotation procedures
- Hardware security modules (HSM)
- Certificate chains
- Audit logging
- Recovery mechanisms

## Future Enhancements

Potential additions:
1. Multi-signature transactions (require multiple signers)
2. Signature aggregation (combine signatures to save space)
3. Zero-knowledge proofs (prove ownership without revealing key)
4. Key delegation (allow temporary signing authority)
