package blockchain

import (
	"fmt"

	"github.com/dgraph-io/badger/v4"
)

func (bc *Blockchain) GetBlockByHeight(height uint64) (*Block, error) {
	if height == 0 {
		return bc.GetGenesisBlock()
	}

	currentHeight := bc.GetHeight()
	if height > currentHeight {
		return nil, fmt.Errorf("height %d exceeds chain length %d", height, currentHeight)
	}

	var block *Block
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
			block, err = DeserializeBlock(val)
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

func (bc *Blockchain) GetBlock(hash []byte) (*Block, error) {
	var block *Block

	err := bc.DB.View(func(txn *badger.Txn) error {
		item, err := txn.Get(hash)
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			block, err = DeserializeBlock(val)
			return err
		})
	})

	return block, err
}
func (bc *Blockchain) GetGenesisBlock() (*Block, error) {
	var genesis *Block
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

				var block *Block
				err = item.Value(func(blockData []byte) error {
					block, err = DeserializeBlock(blockData)
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
