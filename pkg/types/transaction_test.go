package types_test

// test for transaction.go
import (
	"testing"

	"github.com/karimseh/gochain/pkg/crypto"
	"github.com/karimseh/gochain/pkg/types"
	"github.com/karimseh/gochain/pkg/wallet"
	"github.com/stretchr/testify/assert"
)

func TestTrabsaction_SignAndVerify(t *testing.T) {
	w := wallet.NewWallet()
	pubKeyBytes := crypto.PublicKeyToBytes(w.PublicKey)

	t.Run("Valid Tx", func(t *testing.T) {
		tx := types.NewTransaction(w.Address, "to", 10, 0, pubKeyBytes)
		err := tx.Sign(w)
		assert.NoError(t, err)
		assert.NotEmpty(t, tx.Signature)
		assert.NotEmpty(t, tx.Hash)
		assert.True(t, tx.Verify())
	})

	t.Run("Invalid Tx", func(t *testing.T) {
		tx := types.NewTransaction(w.Address, "to", 10, 0, pubKeyBytes)
		tx.Signature = []byte("invalid")
		assert.False(t, tx.Verify())
	})

	t.Run("Wrong pk", func(t *testing.T) {
		tx := types.NewTransaction(w.Address, "to", 10, 0, []byte("wrong"))
		tx.Sign(w)
		assert.False(t, tx.Verify())
	})
}

func TestTransaction_CalculateHash(t *testing.T) {
	tx1 := types.NewTransaction("A", "B", 100, 1, []byte("pubkey"))
	hash1 := tx1.CalculateHash()

	tx2 := types.NewTransaction("A", "B", 100, 1, []byte("pubkey"))
	hash2 := tx2.CalculateHash()
	assert.Equal(t, hash1, hash2, "Same transactions should have same hash")

	tx3 := types.NewTransaction("A", "B", 100, 2, []byte("pubkey"))
	hash3 := tx3.CalculateHash()
	assert.NotEqual(t, hash1, hash3, "Different amount should change hash")
}

func TestTransaction_VerifyCoinbase(t *testing.T) {
	t.Run("Valid Coinbase", func(t *testing.T) {
		cbTx := types.NewCoinbaseTx("miner")
		assert.True(t, cbTx.Verify())
	})

	t.Run("Invalid Coinbase Amount", func(t *testing.T) {
		cbTx := types.NewCoinbaseTx("miner")
		cbTx.Ammount = 51
		assert.False(t, cbTx.Verify())
	})

	t.Run("Coinbase With Signature", func(t *testing.T) {
		cbTx := types.NewCoinbaseTx("miner")
		cbTx.Signature = []byte("invalid-sig")
		assert.False(t, cbTx.Verify())
	})
}