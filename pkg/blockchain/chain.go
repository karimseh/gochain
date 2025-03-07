package blockchain

import (
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/karimseh/gochain/pkg/types"
)

type Blockchain struct {
	DB       *badger.DB
	LastHash []byte
}

func NewBlockchain() (*Blockchain, error) {
	dbPath := dbPath()
	db, err := openDB(dbPath)
	if err != nil {
		return nil, err
	}
	bc := &Blockchain{DB: db}
	// Load last hash immediately after DB connection
	err = bc.DB.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lastHash"))
		if err != nil {
			return err
		}
		return item.Value(func(val []byte) error {
			bc.LastHash = append([]byte{}, val...)
			return nil
		})
	})

	if err == badger.ErrKeyNotFound {
		if err := bc.initialize(); err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	return bc, nil
}

func (bc *Blockchain) initialize() error {
	return bc.DB.Update(func(txn *badger.Txn) error {
		genesis := createGenesisBlock()
		if err := genesis.Validate(); err != nil {
			return err
		}

		if err := txn.Set(genesis.Hash, genesis.Serialize()); err != nil {
			return err
		}
		if err := txn.Set([]byte("lastHash"), genesis.Hash); err != nil {
			return err
		}
		bc.LastHash = genesis.Hash
		return nil
	})
}

func createGenesisBlock() *types.Block {
	genesis := &types.Block{
		Index:     0,
		Timestamp: time.Now().Unix(),
		Data:      []byte("GENESIS BLOCK"),
		PrevHash:  []byte{},
	}
	// Calculate hash for genesis block
	genesis.Hash = genesis.CalculateHash()
	return genesis

}
