package blockchain

import (
	"sync"

	"github.com/dgraph-io/badger/v4"
	"github.com/karimseh/gochain/pkg/mempool"
	"github.com/karimseh/gochain/pkg/state"
	"github.com/karimseh/gochain/pkg/types"
)

type Blockchain struct {
	DB       *badger.DB
	State    *state.State
	Mempool  *mempool.TxPool
	LastHash []byte
	height   uint64
	genesis  *types.Block
	mu       sync.RWMutex
}

func NewBlockchain() (*Blockchain, error) {
	dbPath := dbPath()
	db, err := openDB(dbPath)
	if err != nil {
		return nil, err
	}
	bc := &Blockchain{DB: db, State: state.NewState(db), Mempool: mempool.NewTxPool()}

	if err := bc.initialize(); err != nil {
		return nil, err
	}

	return bc, nil
}

func (bc *Blockchain) initialize() error {
	return bc.DB.Update(func(txn *badger.Txn) error {
		_, err := txn.Get([]byte("lastHash"))
		if err == badger.ErrKeyNotFound {
			return bc.createGenesisBlock()
		}
		return bc.loadChainState(txn)
	})
}

func (bc *Blockchain) createGenesisBlock() error {
	genesis := types.NewBlock(0, []*types.Transaction{}, []byte{}, "GENESIS")
	genesis.Hash = genesis.CalculateHash()

	return bc.DB.Update(func(txn *badger.Txn) error {
		if err := txn.Set(genesis.Hash, genesis.Serialize()); err != nil {
			return err
		}
		if err := txn.Set([]byte("lastHash"), genesis.Hash); err != nil {
			return err
		}
		bc.LastHash = genesis.Hash
		bc.genesis = genesis
		bc.height = 0
		return nil
	})
}

func (bc *Blockchain) loadChainState(txn *badger.Txn) error {
	item, err := txn.Get([]byte("lastHash"))
	if err != nil {
		return err
	}
	return item.Value(func(val []byte) error {
		lastBlock, err := bc.GetBlock(val)
		if err != nil {
			return err
		}

		bc.mu.Lock()
		defer bc.mu.Unlock()

		bc.LastHash = lastBlock.Hash
		bc.height = lastBlock.Header.Index
		bc.genesis, _ = bc.GetGenesisBlock()
		return nil
	})
}
