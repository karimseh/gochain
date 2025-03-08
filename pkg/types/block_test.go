package types_test


import (
	"testing"

	"github.com/karimseh/gochain/pkg/crypto"
	"github.com/karimseh/gochain/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBlock(t *testing.T) {
	parentHash := crypto.HashData([]byte("parent"))
	txs := []*types.Transaction{
		dummyTransaction(),
		dummyTransaction(),
	}

	block := types.NewBlock(1, txs, parentHash, "miner-address")

	assert.Equal(t, uint64(1), block.Header.Index)
	assert.Equal(t, "miner-address", block.Header.Miner)
	assert.Equal(t, parentHash, block.Header.ParentHash)
	assert.NotZero(t, block.Header.Timestamp)
	assert.Len(t, block.Transactions, 2)
	assert.NotNil(t, block.MerkleRoot)
	assert.NotNil(t, block.Hash)
}

func TestBlockValidation(t *testing.T) {
	t.Run("Valid Block", func(t *testing.T) {
		block := validBlock()
		assert.NoError(t, block.Validate())
	})

	t.Run("Tampered Hash", func(t *testing.T) {
		block := validBlock()
		block.Hash[0]++ // Tamper with hash
		assert.ErrorContains(t, block.Validate(), "invalid block hash")
	})

}

func TestCalculateHash(t *testing.T) {
	block1 := validBlock()
	block2 := validBlock()

	// Same blocks should have same hash
	assert.Equal(t, block1.Hash, block2.Hash)

	// Different index
	block2.Header.Index = 2
	assert.NotEqual(t, block1.Hash, block2.CalculateHash())

	// Different parent hash
	block2 = validBlock()
	block2.Header.ParentHash = crypto.HashData([]byte("different"))
	assert.NotEqual(t, block1.Hash, block2.CalculateHash())
}

func TestMerkleRoot(t *testing.T) {


	t.Run("Single Transaction", func(t *testing.T) {
		tx := dummyTransaction()
		block := types.NewBlock(0, []*types.Transaction{tx}, nil, "miner")
		assert.Equal(t, tx.Hash, block.MerkleRoot)
	})

	t.Run("Multiple Transactions", func(t *testing.T) {
		txs := []*types.Transaction{
			dummyTransaction(),
			dummyTransaction(),
		}
		block := types.NewBlock(0, txs, nil, "miner")
		expected := crypto.BuildMerkleRoot([][]byte{txs[0].Hash, txs[1].Hash})
		assert.Equal(t, expected, block.MerkleRoot)
	})
}

func TestSerialization(t *testing.T) {
	original := validBlock()

	// Test round-trip serialization
	data := original.Serialize()
	deserialized, err := types.DeserializeBlock(data)
	require.NoError(t, err)

	assert.Equal(t, original.Header, deserialized.Header)
	assert.Equal(t, original.Hash, deserialized.Hash)
	assert.Equal(t, original.MerkleRoot, deserialized.MerkleRoot)
	assert.Len(t, deserialized.Transactions, len(original.Transactions))
}

func TestEdgeCases(t *testing.T) {
	t.Run("Nil Parent Hash", func(t *testing.T) {
		block := types.NewBlock(0, nil, nil, "miner")
		assert.Nil(t, block.Header.ParentHash)
		assert.Error(t, block.Validate())
	})

	t.Run("Max Values", func(t *testing.T) {
		block := types.NewBlock(^uint64(0), nil, make([]byte, 32), "miner")
		block.Header.Nonce = ^uint64(0)
		block.Header.Difficulty = 100
		assert.NotPanics(t, func() { block.CalculateHash() })
	})

	t.Run("Empty Miner Address", func(t *testing.T) {
		block := types.NewBlock(0, nil, nil, "")
		assert.Error(t, block.Validate())
	})
}

func validBlock() *types.Block {
	block := types.NewBlock(1, []*types.Transaction{dummyTransaction()}, make([]byte, 32), "miner")
	block.Hash = block.CalculateHash()
	return block
}

func dummyTransaction() *types.Transaction {
	tx := &types.Transaction{
		From:    "from",
		To:      "to",
		Ammount:  100,
		Nonce:   1,
		PubKey:  []byte("pubkey"),
		Hash:    crypto.HashData([]byte("dummy")),
	}
	return tx
}
