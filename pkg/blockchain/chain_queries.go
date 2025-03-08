package blockchain

import (
	"bytes"
	"fmt"

	"github.com/dgraph-io/badger/v4"
	"github.com/karimseh/gochain/pkg/types"
)

func (bc *Blockchain) GetHeight() uint64 {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	return bc.height
}

func (bc *Blockchain) GetBlock(hash []byte) (*types.Block, error) {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

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

func (bc *Blockchain) GetLastBlock() *types.Block {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	block, _ := bc.GetBlock(bc.LastHash)
	return block
}

func (bc *Blockchain) GetGenesisBlock() (*types.Block, error) {
	if bc.genesis != nil {
		return bc.genesis, nil
	}

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

				if block.Header.Index == 0 {
					genesis = block
					return nil
				}
				currentHash = block.Header.ParentHash
			}
		})
	})
	return genesis, err
}

func (bc *Blockchain) AddBlock(block *types.Block) error {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	if block.Header.Index != bc.height+1 {
		return fmt.Errorf("invalid block index: %d, expected: %d", block.Header.Index, bc.height+1)
	}
	if !bytes.Equal(block.Header.ParentHash, bc.LastHash) {
		return fmt.Errorf("invalid parent hash")
	}

	if err := block.Validate(); err != nil {
		return err
	}

	if err := bc.State.ApplyBlock(block); err != nil {
		return err
	}
	// Update state root
	stateRoot, err := bc.State.CalculateStateRoot()
	if err != nil {
		return err
	}
	block.StateRoot = stateRoot

	err = bc.DB.Update(func(txn *badger.Txn) error {
		if err := txn.Set(block.Hash, block.Serialize()); err != nil {
			return err
		}
		if err := txn.Set([]byte("lastHash"), block.Hash); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	bc.LastHash = block.Hash
	bc.height = block.Header.Index

	bc.Mempool.RemoveTxs(block.Transactions[1:])

	return nil

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

			if block.Header.Index == 0 {
				break // Reached genesis block
			}
			currentHash = block.Header.ParentHash
		}
		return nil
	})
}
