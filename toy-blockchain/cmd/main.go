package main

import (
	"fmt"
	"os"
	"strconv"
	"toy-blockchain/blockchain"
)

func usage() {
	fmt.Println("Usage:")
	fmt.Println("  go run ./cmd init")
	fmt.Println("  go run ./cmd addtx <sender> <recipient> <amount>")
	fmt.Println("  go run ./cmd mine")
	fmt.Println("  go run ./cmd print")
	fmt.Println("  go run ./cmd balances")
	fmt.Println("  go run ./cmd validate")
}

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		usage()
		os.Exit(1)
	}

	cmd := args[0]
	bc, err := blockchain.LoadFromFile(blockchain.DefaultDataFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load blockchain: %v\n", err)
		os.Exit(1)
	}

	switch cmd {
	case "init":
		bc = blockchain.NewBlockchain()
		if err := bc.SaveToFile(blockchain.DefaultDataFile); err != nil {
			fmt.Fprintf(os.Stderr, "failed to save blockchain: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("blockchain initialized")

	case "addtx":
		if len(args) != 4 {
			fmt.Fprintln(os.Stderr, "addtx requires sender, recipient, and amount")
			usage()
			os.Exit(1)
		}
		amount, err := strconv.ParseFloat(args[3], 64)
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid amount: %v\n", err)
			os.Exit(1)
		}
		tx := blockchain.Transaction{Sender: args[1], Recipient: args[2], Amount: amount}
		if err := bc.AddTransaction(tx); err != nil {
			fmt.Fprintf(os.Stderr, "failed to add transaction: %v\n", err)
			os.Exit(1)
		}
		if err := bc.SaveToFile(blockchain.DefaultDataFile); err != nil {
			fmt.Fprintf(os.Stderr, "failed to save blockchain: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("transaction added")

	case "mine":
		bc.MinePendingTransactions()
		if err := bc.SaveToFile(blockchain.DefaultDataFile); err != nil {
			fmt.Fprintf(os.Stderr, "failed to save blockchain: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("pending transactions mined")

	case "print":
		bc.Print()

	case "balances":
		balances := bc.GetBalances()
		for account, balance := range balances {
			fmt.Printf("%s: %.8f\n", account, balance)
		}

	case "validate":
		if err := bc.Validate(); err != nil {
			fmt.Printf("Validation failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Validation passed: blockchain is valid")

	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", cmd)
		usage()
		os.Exit(1)
	}
}
