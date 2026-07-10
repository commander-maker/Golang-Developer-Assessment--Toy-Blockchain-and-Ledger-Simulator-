package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
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
	fmt.Println()
	fmt.Println("Global flags:")
	fmt.Println("  --difficulty <n>   Mining difficulty (default: 4)")
	fmt.Println("  --blocksize <n>   Maximum transactions per block (default: 10)")
	fmt.Println("  --file <path>     Blockchain data file (default: blockchain.json)")
}

func parseCLI(args []string) (string, []string, int, int, string, error) {
	difficulty := blockchain.DefaultDifficulty
	blockSize := blockchain.DefaultBlockSize
	dataFile := blockchain.DefaultDataFile

	remaining := []string{}
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if strings.HasPrefix(arg, "--difficulty=") {
			value := strings.TrimPrefix(arg, "--difficulty=")
			d, err := strconv.Atoi(value)
			if err != nil {
				return "", nil, 0, 0, "", fmt.Errorf("invalid difficulty: %v", err)
			}
			difficulty = d
			continue
		}
		if strings.HasPrefix(arg, "--blocksize=") {
			value := strings.TrimPrefix(arg, "--blocksize=")
			b, err := strconv.Atoi(value)
			if err != nil {
				return "", nil, 0, 0, "", fmt.Errorf("invalid blocksize: %v", err)
			}
			blockSize = b
			continue
		}
		if strings.HasPrefix(arg, "--file=") {
			dataFile = strings.TrimPrefix(arg, "--file=")
			continue
		}

		switch arg {
		case "--difficulty":
			if i+1 >= len(args) {
				return "", nil, 0, 0, "", fmt.Errorf("missing value for --difficulty")
			}
			d, err := strconv.Atoi(args[i+1])
			if err != nil {
				return "", nil, 0, 0, "", fmt.Errorf("invalid difficulty: %v", err)
			}
			difficulty = d
			i++
		case "--blocksize":
			if i+1 >= len(args) {
				return "", nil, 0, 0, "", fmt.Errorf("missing value for --blocksize")
			}
			b, err := strconv.Atoi(args[i+1])
			if err != nil {
				return "", nil, 0, 0, "", fmt.Errorf("invalid blocksize: %v", err)
			}
			blockSize = b
			i++
		case "--file":
			if i+1 >= len(args) {
				return "", nil, 0, 0, "", fmt.Errorf("missing value for --file")
			}
			dataFile = args[i+1]
			i++
		default:
			remaining = append(remaining, arg)
		}
	}

	if len(remaining) == 0 {
		return "", nil, 0, 0, "", fmt.Errorf("command required")
	}

	cmd := remaining[0]
	if !isCommand(cmd) {
		return "", nil, 0, 0, "", fmt.Errorf("unknown command: %s", cmd)
	}

	return cmd, remaining[1:], difficulty, blockSize, dataFile, nil
}

func isCommand(arg string) bool {
	switch arg {
	case "init", "addtx", "mine", "print", "balances", "validate":
		return true
	default:
		return false
	}
}

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		usage()
		os.Exit(1)
	}

	cmd, cmdArgs, difficulty, blockSize, dataFile, err := parseCLI(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to parse CLI flags: %v\n", err)
		os.Exit(1)
	}
	if cmd == "" {
		usage()
		os.Exit(1)
	}

	bc, err := blockchain.LoadFromFile(dataFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load blockchain: %v\n", err)
		os.Exit(1)
	}
	bc.Difficulty = difficulty
	bc.BlockSize = blockSize

	switch cmd {
	case "init":
		bc = blockchain.NewBlockchainWithConfig(difficulty, blockSize)
		if err := bc.SaveToFile(dataFile); err != nil {
			fmt.Fprintf(os.Stderr, "failed to save blockchain: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("blockchain initialized")

	case "addtx":
		if len(cmdArgs) != 3 {
			fmt.Fprintln(os.Stderr, "addtx requires sender, recipient, and amount")
			usage()
			os.Exit(1)
		}
		amount, err := strconv.ParseFloat(cmdArgs[2], 64)
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid amount: %v\n", err)
			os.Exit(1)
		}
		tx := blockchain.Transaction{Sender: cmdArgs[0], Recipient: cmdArgs[1], Amount: amount}
		if err := bc.AddTransaction(tx); err != nil {
			fmt.Fprintf(os.Stderr, "failed to add transaction: %v\n", err)
			os.Exit(1)
		}
		if err := bc.SaveToFile(dataFile); err != nil {
			fmt.Fprintf(os.Stderr, "failed to save blockchain: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("transaction added")

	case "mine":
		bc.MinePendingTransactions()
		if err := bc.SaveToFile(dataFile); err != nil {
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
