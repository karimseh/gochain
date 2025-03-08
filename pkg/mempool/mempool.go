package mempool

import (
	"encoding/hex"
	"fmt"
	"sync"

	"github.com/karimseh/gochain/pkg/types"
)

type TxPool struct {
	mu           sync.RWMutex
	transactions map[string]*types.Transaction
	orderedTxs   []*types.Transaction
}

func NewTxPool() *TxPool {
	return &TxPool{
		transactions: make(map[string]*types.Transaction),
		orderedTxs:   make([]*types.Transaction, 0),
	}
}

func (pool *TxPool) AddTx(tx *types.Transaction) error {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	txhash := hex.EncodeToString(tx.Hash)

	if _, exists := pool.transactions[txhash]; exists {
		return fmt.Errorf("transaction already exists in pool")
	}

	if tx.Ammount == 0 {
		return fmt.Errorf("transaction ammount must be greater than 0")
	}

	if !tx.Verify() {
		return fmt.Errorf("transaction signature verification failed")
	}

	pool.transactions[txhash] = tx
	pool.orderedTxs = append(pool.orderedTxs, tx)
	return nil
}

func (pool *TxPool) GetTxs(max int) []*types.Transaction {
	pool.mu.RLock()
	defer pool.mu.RUnlock()
	if max <= 0 || max > len(pool.orderedTxs) {
		max = len(pool.orderedTxs)
	}
	return pool.orderedTxs[:max]
}

func (pool *TxPool) RemoveTxs(txs []*types.Transaction) {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	for _, tx := range txs {
		txHash := hex.EncodeToString(tx.Hash)
		delete(pool.transactions, txHash)
	}

	newOrdered := make([]*types.Transaction, 0, len(pool.transactions))
	for _, tx := range pool.orderedTxs {
		if _, exists := pool.transactions[hex.EncodeToString(tx.Hash)]; exists {
			newOrdered = append(newOrdered, tx)
		}
	}
	pool.orderedTxs = newOrdered
}

func (pool *TxPool) PendingCount() int {
	pool.mu.RLock()
	defer pool.mu.RUnlock()
	return len(pool.transactions)
}
