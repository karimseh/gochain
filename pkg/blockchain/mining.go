package blockchain

import (
	"github.com/karimseh/gochain/pkg/consensus"

	"github.com/karimseh/gochain/pkg/types"
)

func (bc *Blockchain) MineBlock(miner string) error {
	// Get transactions from mempool (excluding coinbase)
	txs := bc.Mempool.GetTxs(50)

	// Create coinbase transaction
	coinbase := types.NewCoinbaseTx(miner)
	txs = append([]*types.Transaction{coinbase}, txs...)

	lastBlock := bc.GetLastBlock()
	newBlock := types.NewBlock(
		lastBlock.Header.Index+1,	
		txs,
		lastBlock.Hash,
		miner,
	)

	// Run Proof-of-Work
	pow := consensus.NewProofOfWork(newBlock)
	nonce, hash := pow.Run()
	newBlock.Header.Nonce = nonce
	newBlock.Hash = hash

	return bc.AddBlock(newBlock)
}
