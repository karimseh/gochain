package blockchain

import (
	"fmt"

	"github.com/karimseh/gochain/pkg/consensus"
	"github.com/karimseh/gochain/pkg/types"
)

func (bc *Blockchain) MineBlock(data []byte) error {
	lastBlock, err := bc.GetBlock(bc.LastHash)
	if err != nil {
		return err
	}

	newBlock := types.NewBlock(
		lastBlock.Index + 1,
		data,
		lastBlock.Hash,
		consensus.TargetBits,
	)

	pow := consensus.NewProofOfWork(newBlock)
	nonce, hash := pow.Run()
	
	newBlock.Nonce = nonce
	newBlock.Hash = hash

	if err := newBlock.Validate(); err != nil {
		return fmt.Errorf("mined block validation failed: %v", err)
	}

	return bc.AddBlock(newBlock)
}