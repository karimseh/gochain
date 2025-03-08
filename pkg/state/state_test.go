package state_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dgraph-io/badger/v4"
	"github.com/karimseh/gochain/pkg/crypto"
	"github.com/karimseh/gochain/pkg/state"
	"github.com/karimseh/gochain/pkg/types"
	"github.com/karimseh/gochain/pkg/wallet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupState(t *testing.T) (*state.State, func()) {
	dir := filepath.Join(os.TempDir(), "badger-test", t.Name())
	_ = os.RemoveAll(dir)

	opts := badger.DefaultOptions(dir)
	opts.Logger = nil // Disable logger for tests

	db, err := badger.Open(opts)
	require.NoError(t, err)

	return state.NewState(db), func() {
		db.Close()
		os.RemoveAll(dir)
	}
}

func TestNewState(t *testing.T) {
	s, cleanup := setupState(t)
	defer cleanup()

	assert.NotNil(t, s, "State should be initialized")
	balance, err := s.GetBalance("non-existent")
	assert.NoError(t, err)
	assert.Equal(t, uint64(0), balance, "New state should have zero balance for new accounts")
}

func TestGetAccount(t *testing.T) {
	s, cleanup := setupState(t)
	defer cleanup()

	t.Run("Cached Account", func(t *testing.T) {
		acc := &types.Account{Address: "test", Balance: 100, Nonce: 1}
		require.NoError(t, s.SaveAccount(acc))

		// First get (from DB)

		a1, err := s.GetAccount("test")
		assert.NoError(t, err)
		assert.Equal(t, uint64(100), a1.Balance)

		// Second get (from cache)
		a2, err := s.GetAccount("test")
		assert.NoError(t, err)
		assert.True(t, a1 == a2, "Should return same pointer from cache")
	})

	t.Run("Non-Existent Account", func(t *testing.T) {
		acc, err := s.GetAccount("unknown")
		assert.NoError(t, err)
		assert.Equal(t, "unknown", acc.Address)
		assert.Equal(t, uint64(0), acc.Balance)
		assert.Equal(t, uint64(0), acc.Nonce)
	})
}

func TestGetBalance(t *testing.T) {
	s, cleanup := setupState(t)
	defer cleanup()

	require.NoError(t, s.SaveAccount(&types.Account{
		Address: "balance-test",
		Balance: 1000,
		Nonce:   5,
	}))

	balance, err := s.GetBalance("balance-test")
	assert.NoError(t, err)
	assert.Equal(t, uint64(1000), balance)
}

func TestValidateTx(t *testing.T) {
	s, cleanup := setupState(t)
	defer cleanup()

	w := wallet.NewWallet()
	require.NoError(t, s.SaveAccount(&types.Account{
		Address: w.Address,
		Balance: 100,
		Nonce:   0,
	}))

	tx := types.NewTransaction(w.Address, "recipient", 50, 1, crypto.PublicKeyToBytes(w.PublicKey))
	require.NoError(t, tx.Sign(w))

	t.Run("Valid Transaction", func(t *testing.T) {
		assert.NoError(t, s.ValidateTx(tx))
	})

	t.Run("Invalid Nonce", func(t *testing.T) {
		badTx := *tx
		badTx.Nonce = 2
		badTx.Sign(w)
		assert.ErrorContains(t, s.ValidateTx(&badTx), "invalid nonce")
	})

	t.Run("Insufficient Balance", func(t *testing.T) {
		badTx := *tx
		badTx.Ammount = 200
		badTx.Sign(w)
		assert.ErrorContains(t, s.ValidateTx(&badTx), "insufficient balance")
	})
	t.Run("Invalid Hash", func(t *testing.T) {
		badTx := *tx
		badTx.Ammount = 1
		assert.Error(t, s.ValidateTx(&badTx))
	})

	t.Run("Invalid Signature", func(t *testing.T) {
		badTx := *tx
		badTx.Signature[0]++
		assert.ErrorContains(t, s.ValidateTx(&badTx), "signature verification")
	})
}

func TestApplyTx(t *testing.T) {
	s, cleanup := setupState(t)
	defer cleanup()

	w := wallet.NewWallet()
	require.NoError(t, s.SaveAccount(&types.Account{
		Address: w.Address,
		Balance: 200,
		Nonce:   0,
	}))

	tx := types.NewTransaction(w.Address, "recipient", 50, 1, crypto.PublicKeyToBytes(w.PublicKey))
	require.NoError(t, tx.Sign(w))

	t.Run("Valid Transaction", func(t *testing.T) {
		err := s.ApplyTx(tx)
		assert.NoError(t, err)

		sender, _ := s.GetAccount(w.Address)
		assert.Equal(t, uint64(150), sender.Balance)
		assert.Equal(t, uint64(1), sender.Nonce)

		receiver, _ := s.GetAccount("recipient")
		assert.Equal(t, uint64(50), receiver.Balance)
	})

	t.Run("Invalid Transaction", func(t *testing.T) {
		badTx := *tx
		badTx.Ammount = 300
		err := s.ApplyTx(&badTx)
		assert.Error(t, err)
	})
}

func TestApplyBlock(t *testing.T) {
	s, cleanup := setupState(t)
	defer cleanup()

	minerWallet := wallet.NewWallet()
	coinbase := types.NewCoinbaseTx(minerWallet.Address)

	w := wallet.NewWallet()
	require.NoError(t, s.SaveAccount(&types.Account{
		Address: w.Address,
		Balance: 1000,
		Nonce:   0,
	}))

	tx := types.NewTransaction(w.Address, "recipient", 200, 1, crypto.PublicKeyToBytes(w.PublicKey))
	require.NoError(t, tx.Sign(w))

	block := types.NewBlock(1, []*types.Transaction{coinbase, tx}, make([]byte, 32), minerWallet.Address)

	err := s.ApplyBlock(block)
	assert.NoError(t, err)

	t.Run("Coinbase Applied", func(t *testing.T) {
		minerAcc, _ := s.GetAccount(minerWallet.Address)
		assert.Equal(t, uint64(types.CoinbaseAmount), minerAcc.Balance)
	})

	t.Run("Transaction Applied", func(t *testing.T) {
		sender, _ := s.GetAccount(w.Address)
		assert.Equal(t, uint64(800), sender.Balance)

		receiver, _ := s.GetAccount("recipient")
		assert.Equal(t, uint64(200), receiver.Balance)
	})
}

func TestCalculateStateRoot(t *testing.T) {
	s, cleanup := setupState(t)
	defer cleanup()

	// Create test accounts
	accounts := []*types.Account{
		{Address: "a", Balance: 100, Nonce: 1},
		{Address: "b", Balance: 200, Nonce: 2},
		{Address: "c", Balance: 300, Nonce: 3},
	}

	for _, acc := range accounts {
		require.NoError(t, s.SaveAccount(acc))
	}

	root1, err := s.CalculateStateRoot()
	assert.NoError(t, err)
	assert.NotNil(t, root1)

	// Modify an account and check root changes
	modified := *accounts[0]
	modified.Balance = 150
	require.NoError(t, s.SaveAccount(&modified))

	root2, err := s.CalculateStateRoot()
	assert.NoError(t, err)
	assert.NotEqual(t, root1, root2)
}

func TestEdgeCases(t *testing.T) {
	s, cleanup := setupState(t)
	defer cleanup()

	t.Run("Zero Amount Transaction", func(t *testing.T) {
		w := wallet.NewWallet()
		s.SaveAccount(&types.Account{Address: w.Address, Balance: 100})

		tx := types.NewTransaction(w.Address, "empty", 0, 1, crypto.PublicKeyToBytes(w.PublicKey))
		tx.Sign(w)

		err := s.ApplyTx(tx)
		assert.Error(t, err)
	})

}
