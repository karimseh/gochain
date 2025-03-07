package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/karimseh/gochain/pkg/blockchain"
	"github.com/karimseh/gochain/pkg/types"
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
	case "mine":
		if len(os.Args) < 3 {
			log.Fatal("Usage: mine <data>")
		}
		data := os.Args[2]
		err := bc.MineBlock([]byte(data))
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("New block mined!")
		handleStatus(bc)
	case "printchain":
		handlePrintChain(bc)
	default:
		printUsage()

	}

}

func handleStatus(bc *blockchain.Blockchain) {
	height := bc.GetHeight()
	lastBlock, err := bc.GetBlock(bc.LastHash)
	if err != nil {
		log.Fatal(err)
	}

	genesis, err := bc.GetBlockByHeight(0)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("=== Blockchain Status ===")
	fmt.Printf("Chain Height: %d\n", height)
	fmt.Printf("Genesis Hash: %x\n", genesis.Hash)
	fmt.Printf("Last Block Hash: %x\n", lastBlock.Hash)
	fmt.Printf("Last Block Nonce: %d\n", lastBlock.Nonce)
	fmt.Printf("Difficulty: %d\n", lastBlock.Difficulty)
}
func handlePrintChain(bc *blockchain.Blockchain) {
	fmt.Println("\n=== Blockchain Contents ===")

	err := bc.IterateBlocks(func(block *types.Block) error {
		printBlock(block)
		return nil
	})

	if err != nil {
		log.Fatalf("Error reading chain: %v", err)
	}
}
func printBlock(block *types.Block) {
	fmt.Printf("\nBlock %d\n", block.Index)
	fmt.Println("---------------------")
	fmt.Printf("Timestamp:  %s\n", time.Unix(block.Timestamp, 0).Format(time.RFC3339))
	fmt.Printf("Data:       %s\n", string(block.Data))
	fmt.Printf("Prev Hash:  %x\n", block.PrevHash)
	fmt.Printf("Hash:       %x\n", block.Hash)
	fmt.Printf("Nonce:      %d\n", block.Nonce)
	fmt.Printf("Difficulty: %d\n", block.Difficulty)
	fmt.Println("---------------------")
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  status      - Show blockchain status")
	fmt.Println("  mine <data> - Mine new block with data")
	fmt.Println("  printchain  - Print all blocks")

}
