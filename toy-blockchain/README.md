# 🔗 Toy Blockchain — Built in Go

A ground-up blockchain implementation in Go, built step by step.

## Day 1 Objectives

- [x] Project structure
- [x] `go.mod`
- [x] `Block` struct
- [x] `Blockchain` struct
- [x] Genesis Block
- [x] SHA-256 deterministic hashing
- [x] Print the blockchain
- [x] Basic unit tests for hashing

## Project Structure

```
toy-blockchain/
│
├── cmd/
│   └── main.go          ← Entry point
│
├── blockchain/
│   ├── block.go         ← Block struct & constructor
│   ├── blockchain.go    ← Blockchain struct, Genesis, AddBlock, Print
│   └── hash.go          ← SHA-256 hashing logic
│
├── tests/
│   └── hash_test.go     ← Unit tests for hashing
│
├── go.mod
└── README.md
```

## How to Run

```bash
go run cmd/main.go
```

## How to Test

```bash
go test ./tests/...
```
