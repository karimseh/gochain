package blockchain

import (
	"bytes"
	"fmt"

	"github.com/dgraph-io/badger/v4"
	"github.com/karimseh/gochain/pkg/types"
)

func (bc *Blockchain) GetBlockByHeight(height uint64) (*types.Block, error) {
	if height == 0 {
		return bc.GetGenesisBlock()
	}

	currentHeight := bc.GetHeight()
	if height > currentHeight {
		return nil, fmt.Errorf("height %d exceeds chain length %d", height, currentHeight)
	}

	var block *types.Block
	err := bc.DB.View(func(txn *badger.Txn) error {
		currentHash := bc.LastHash

		for currentHeight > height {
			currentBlock, err := bc.GetBlock(currentHash)
			if err != nil {
				return err
			}
			currentHash = currentBlock.PrevHash
			currentHeight--
		}

		item, err := txn.Get(currentHash)
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			block, err = types.DeserializeBlock(val)
			return err
		})
	})
	return block, err
}
func (bc *Blockchain) GetHeight() uint64 {
	var height uint64

	_ = bc.DB.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lastHash"))
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			lastBlock, err := bc.GetBlock(val)
			if err != nil {
				return err
			}
			height = lastBlock.Index
			return nil
		})
	})

	return height
}

func (bc *Blockchain) GetBlock(hash []byte) (*types.Block, error) {
	var block *types.Block

	err := bc.DB.View(func(txn *badger.Txn) error {
		item, err := txn.Get(hash)
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			block, err = types.DeserializeBlock(val)
			return err
		})
	})

	return block, err
}
func (bc *Blockchain) GetGenesisBlock() (*types.Block, error) {
	var genesis *types.Block
	err := bc.DB.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lastHash"))
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			// Follow the chain back to genesis
			currentHash := val
			for {
				item, err := txn.Get(currentHash)
				if err != nil {
					return err
				}

				var block *types.Block
				err = item.Value(func(blockData []byte) error {
					block, err = types.DeserializeBlock(blockData)
					return err
				})

				if block.Index == 0 {
					genesis = block
					return nil
				}
				currentHash = block.PrevHash
			}
		})
	})
	return genesis, err
}

func (bc *Blockchain) AddBlock(block *types.Block) error {
	return bc.DB.Update(func(txn *badger.Txn) error {
		if err := block.Validate(); err != nil {
			return err
		}
		
		if !bytes.Equal(block.PrevHash, bc.LastHash) {
			return fmt.Errorf("invalid previous hash")
		}
		
		// Store the block
		if err := txn.Set(block.Hash, block.Serialize()); err != nil {
			return err
		}
		
		// Update last hash
		if err := txn.Set([]byte("lastHash"), block.Hash); err != nil {
			return err
		}
		bc.LastHash = block.Hash
		return nil
	})
}

func (bc *Blockchain) IterateBlocks(handler func(*types.Block) error) error {
	return bc.DB.View(func(txn *badger.Txn) error {
		currentHash := bc.LastHash
		
		for {
			item, err := txn.Get(currentHash)
			if err != nil {
				return err
			}
			
			var block *types.Block
			err = item.Value(func(val []byte) error {
				block, err = types.DeserializeBlock(val)
				return err
			})
			
			if err != nil {
				return err
			}
			
			if err := handler(block); err != nil {
				return err
			}
			
			if block.Index == 0 {
				break // Reached genesis block
			}
			currentHash = block.PrevHash
		}
		return nil
	})
}