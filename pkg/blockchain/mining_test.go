package blockchain_test

import (
	"testing"

	"github.com/karimseh/gochain/pkg/blockchain"
	"github.com/karimseh/gochain/pkg/crypto"
	"github.com/karimseh/gochain/pkg/types"
	"github.com/karimseh/gochain/pkg/wallet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupBlockchain(t *testing.T) (*blockchain.Blockchain, func()) {

	bc, err := blockchain.NewBlockchain()
	require.NoError(t, err)

	return bc, func() {
		_ = bc.CloseDB()
	}
}

func TestMineBlock(t *testing.T) {
	t.Run("Valid Block Mining", func(t *testing.T) {
		bc, cleanup := setupBlockchain(t)
		defer cleanup()
		minerWallet := wallet.NewWallet()

		// Add some transactions to mempool
		tx1 := createValidTransaction(t, bc, 100)
		tx2 := createValidTransaction(t, bc, 50)
		require.NoError(t, bc.Mempool.AddTx(tx1))
		require.NoError(t, bc.Mempool.AddTx(tx2))

		err := bc.MineBlock(minerWallet.Address)
		require.NoError(t, err)

		// Verify block was added
		lastBlock := bc.GetLastBlock()
		assert.Equal(t, uint64(1), lastBlock.Header.Index)
		assert.Len(t, lastBlock.Transactions, 3) // Coinbase + 2 transactions
		assert.Equal(t, minerWallet.Address, lastBlock.Header.Miner)

		// Verify Proof-of-Work
		assert.True(t, crypto.ValidateHash(lastBlock.Hash, lastBlock.Header.Difficulty))
	})

	t.Run("Mining Empty Block", func(t *testing.T) {
		bc, cleanup := setupBlockchain(t)
		defer cleanup()

		err := bc.MineBlock("miner")
		require.NoError(t, err)

		lastBlock := bc.GetLastBlock()
		assert.Len(t, lastBlock.Transactions, 1) // Just coinbase
	})

	t.Run("Coinbase Transaction Included", func(t *testing.T) {
		bc, cleanup := setupBlockchain(t)
		defer cleanup()
		minerAddr := "miner-address"

		err := bc.MineBlock(minerAddr)
		require.NoError(t, err)

		lastBlock := bc.GetLastBlock()
		coinbase := lastBlock.Transactions[0]
		assert.True(t, coinbase.IsCoinbase())
		assert.Equal(t, minerAddr, coinbase.To)
		assert.Equal(t, uint64(types.CoinbaseAmount), coinbase.Ammount)
	})

	t.Run("Transaction Pool Cleared", func(t *testing.T) {
		bc, cleanup := setupBlockchain(t)
		defer cleanup()

		tx := createValidTransaction(t, bc, 100)
		require.NoError(t, bc.Mempool.AddTx(tx))
		require.Equal(t, 1, bc.Mempool.PendingCount())

		require.NoError(t, bc.MineBlock("miner"))

		assert.Equal(t, 0, bc.Mempool.PendingCount())
	})
}

func createValidTransaction(t *testing.T, bc *blockchain.Blockchain, amount uint64) *types.Transaction {
	senderWallet := wallet.NewWallet()
	receiverWallet := wallet.NewWallet()

	// Fund sender account
	require.NoError(t, bc.State.SaveAccount(&types.Account{
		Address: senderWallet.Address,
		Balance: 1000,
		Nonce:   0,
	}))

	tx := types.NewTransaction(
		senderWallet.Address,
		receiverWallet.Address,
		amount,
		1, // Nonce
		crypto.PublicKeyToBytes(senderWallet.PublicKey),
	)
	require.NoError(t, tx.Sign(senderWallet))

	return tx
}
