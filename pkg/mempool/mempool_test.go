package mempool_test

import (
	"testing"
	"sync"
	"time"

	"github.com/karimseh/gochain/pkg/crypto"
	"github.com/karimseh/gochain/pkg/mempool"
	"github.com/karimseh/gochain/pkg/types"
	"github.com/karimseh/gochain/pkg/wallet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newValidTx() *types.Transaction {
	w := wallet.NewWallet()
	tx := types.NewTransaction(w.Address, "recipient", 100, 1, crypto.PublicKeyToBytes(w.PublicKey))
	tx.Sign(w)
	return tx
}

func TestTxPool_AddTx(t *testing.T) {
	pool := mempool.NewTxPool()

	t.Run("Valid Transaction", func(t *testing.T) {
		tx := newValidTx()
		err := pool.AddTx(tx)
		assert.NoError(t, err)
		assert.Equal(t, 1, pool.PendingCount())
	})

	t.Run("Duplicate Transaction", func(t *testing.T) {
		tx := newValidTx()
		pool.AddTx(tx)
		err := pool.AddTx(tx)
		assert.ErrorContains(t, err, "already exists")
	})

	t.Run("Zero Amount", func(t *testing.T) {
		tx := newValidTx()
		tx.Ammount = 0
		err := pool.AddTx(tx)
		assert.ErrorContains(t, err, "greater than 0")
	})

	t.Run("Invalid Signature", func(t *testing.T) {
		tx := newValidTx()
		tx.Signature[0]++
		err := pool.AddTx(tx)
		assert.ErrorContains(t, err, "signature verification")
	})
}

func TestTxPool_GetTxs(t *testing.T) {
	pool := mempool.NewTxPool()

	txs := []*types.Transaction{
		newValidTx(),
		newValidTx(),
		newValidTx(),
	}

	for _, tx := range txs {
		require.NoError(t, pool.AddTx(tx))
	}

	t.Run("Get All", func(t *testing.T) {
		result := pool.GetTxs(0)
		assert.Len(t, result, 3)
		assert.Equal(t, txs, result)
	})

	t.Run("Get Subset", func(t *testing.T) {
		result := pool.GetTxs(2)
		assert.Len(t, result, 2)
		assert.Equal(t, txs[:2], result)
	})

	t.Run("Max Larger Than Pool", func(t *testing.T) {
		result := pool.GetTxs(5)
		assert.Len(t, result, 3)
	})
}

func TestTxPool_RemoveTxs(t *testing.T) {
	pool := mempool.NewTxPool()

	txs := []*types.Transaction{
		newValidTx(),
		newValidTx(),
		newValidTx(),
	}
	
	for _, tx := range txs {
		require.NoError(t, pool.AddTx(tx))
	}

	// Remove middle transaction
	pool.RemoveTxs([]*types.Transaction{txs[1]})

	assert.Equal(t, 2, pool.PendingCount())
	result := pool.GetTxs(0)
	assert.Len(t, result, 2)
	assert.Contains(t, result, txs[0])
	assert.Contains(t, result, txs[2])
}

func TestTxPool_Concurrency(t *testing.T) {
	pool := mempool.NewTxPool()
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			tx := newValidTx()
			pool.AddTx(tx)
			pool.GetTxs(1)
			pool.PendingCount()
			pool.RemoveTxs([]*types.Transaction{tx})
		}()
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Timeout waiting for concurrency test")
	}
}

func TestTxPool_Ordering(t *testing.T) {
	pool := mempool.NewTxPool()

	txs := make([]*types.Transaction, 5)
	for i := range txs {
		tx := newValidTx()
		txs[i] = tx
		require.NoError(t, pool.AddTx(tx))
	}

	// Remove middle transaction
	pool.RemoveTxs([]*types.Transaction{txs[2]})

	result := pool.GetTxs(0)
	assert.Len(t, result, 4)
	assert.Equal(t, []*types.Transaction{txs[0], txs[1], txs[3], txs[4]}, result)
}

func TestTxPool_EdgeCases(t *testing.T) {
	t.Run("Empty Pool", func(t *testing.T) {
		pool := mempool.NewTxPool()
		assert.Zero(t, pool.PendingCount())
		assert.Empty(t, pool.GetTxs(0))
	})

	t.Run("Remove Non-Existent", func(t *testing.T) {
		pool := mempool.NewTxPool()
		pool.RemoveTxs([]*types.Transaction{newValidTx()})
		assert.Zero(t, pool.PendingCount())
	})
}