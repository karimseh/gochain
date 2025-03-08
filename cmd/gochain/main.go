package main

import (
	"fmt"
	"log"
	"os"

	"github.com/karimseh/gochain/pkg/blockchain"
	"github.com/karimseh/gochain/pkg/types"
	"github.com/karimseh/gochain/pkg/wallet"
)

var bc *blockchain.Blockchain

func main() {
	var err error
	bc, err = blockchain.NewBlockchain()
	if err != nil {
		log.Fatalf("Failed to initialize blockchain: %v", err)
	}
	defer bc.CloseDB()

	if len(os.Args) < 2 {
		printUsage()
		return
	}

	// Handle commands that need continuous operation
	switch os.Args[1] {
	case "createwallet":
		handleCreateWallet()

	case "balance":
		handleBalance()
	case "status":
		handleStatus()
	case "printchain":
		handlePrintChain()
	default:
		printUsage()
	}
}

func handleCreateWallet() {
	w := wallet.NewWallet()

	if err := w.SaveToFile(); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("New wallet created:\nAddress: %s\n", w.Address)
}

func handleBalance() {
	if len(os.Args) < 3 {
		log.Fatal("Usage: balance <address>")
	}
	address := os.Args[2]

	balance, err := bc.State.GetBalance(address)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Balance for %s: %d\n", address, balance)
}

func handleStatus() {
	lastBlock := bc.GetLastBlock()
	pending := bc.Mempool.PendingCount()

	fmt.Println("=== Blockchain Status ===")
	fmt.Printf("Chain Height: %d\n", bc.GetHeight())
	fmt.Printf("Last Block Hash: %x\n", lastBlock.Hash)
	fmt.Printf("Pending Transactions: %d\n", pending)
}

func handlePrintChain() {
	fmt.Println("=== Blockchain Contents ===")
	err := bc.IterateBlocks(func(block *types.Block) error {
		fmt.Printf("\nBlock %d\n", block.Header.Index)
		fmt.Println("---------------------")
		fmt.Printf("Hash: %x\n", block.Hash)
		fmt.Printf("Miner: %s\n", block.Header.Miner)
		fmt.Printf("Transactions: %d\n", len(block.Transactions))
		fmt.Println("---------------------")
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
}

func printUsage() {
	fmt.Println("GoChain CLI - Account-Based Blockchain")
	fmt.Println("Usage:")
	fmt.Println("  createwallet          - Generate new wallet")
	fmt.Println("  balance <address>     - Check account balance")
	fmt.Println("  status                - Show blockchain status")
	fmt.Println("  printchain            - Display all blocks")
}
