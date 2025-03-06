package main

import (
	"fmt"
	"log"
	"os"

	"github.com/karimseh/gochain/pkg/blockchain"
)

func main() {
	bc, err := blockchain.NewBlockchain()
	if err != nil {
		log.Fatalf("Failed to initialize blockchain: %v", err)
	}
	defer bc.CloseDB()

	if len(os.Args) < 2 {
		printUsage()
		return
	}

	switch os.Args[1] {
	case "status":
		handleStatus(bc)
	default:
		printUsage()

	}

}

func handleStatus(bc *blockchain.Blockchain) {

	genesis, err := bc.GetBlockByHeight(0)
	if err != nil {
		log.Fatalf("Failed to get genesis block: %v", err)
	}
	fmt.Println("=== Blockchain Status ===")
	fmt.Printf("Chain Height: %d\n", genesis.Index)
	fmt.Printf("Genesis Hash: %x\n", genesis.Hash)
	fmt.Printf("Last Block Hash: %x\n", bc.LastHash)

}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  status    - Show blockchain status")
}
