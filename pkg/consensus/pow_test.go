package consensus_test

import (
	"testing"
	"time"

	"github.com/karimseh/gochain/pkg/consensus"
	"github.com/karimseh/gochain/pkg/crypto"
	"github.com/karimseh/gochain/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestProofOfWork_Run(t *testing.T) {
	t.Run("Valid Proof of Work", func(t *testing.T) {
		block := createTestBlock(1)
		pow := consensus.NewProofOfWork(block)

		nonce, hash := pow.Run()

		assert.True(t, crypto.ValidateHash(hash, block.Header.Difficulty),
			"Generated hash doesn't meet difficulty requirements")

		// Verify block validation passes
		block.Header.Nonce = nonce
		block.Hash = hash
		assert.NoError(t, block.Validate())
	})

	t.Run("Different Blocks Different Nonce", func(t *testing.T) {
		block1 := createTestBlock(1)
		block2 := createTestBlock(2) // Different index

		pow1 := consensus.NewProofOfWork(block1)
		pow2 := consensus.NewProofOfWork(block2)

		nonce1, hash1 := pow1.Run()
		nonce2, hash2 := pow2.Run()

		assert.NotEqual(t, nonce1, nonce2, "Same nonce for different blocks")
		assert.NotEqual(t, hash1, hash2, "Same hash for different blocks")
	})

	t.Run("Zero Transactions Block", func(t *testing.T) {
		block := &types.Block{
			Header: types.BlockHeader{
				Index:      1,
				Timestamp:  time.Now().Unix(),
				ParentHash: crypto.HashData([]byte("genesis")),
				Miner:      "miner",
			},
			Transactions: []*types.Transaction{},
			MerkleRoot:   crypto.HashData(nil),
		}

		pow := consensus.NewProofOfWork(block)
		nonce, hash := pow.Run()

		assert.True(t, crypto.ValidateHash(hash, block.Header.Difficulty))
		assert.NotZero(t, nonce)
	})

	t.Run("Difficulty Validation", func(t *testing.T) {
		block := createTestBlock(1)
		pow := consensus.NewProofOfWork(block)
		_, hash := pow.Run()

		// Test with higher difficulty than required
		assert.False(t, crypto.ValidateHash(hash, block.Header.Difficulty+10),
			"Hash should not satisfy higher difficulty")
	})
}

func TestProofOfWork_EdgeCases(t *testing.T) {
	t.Run("Already Valid Block", func(t *testing.T) {
		block := createTestBlock(1)

		// Manually find a valid nonce
		pow := consensus.NewProofOfWork(block)
		validNonce, validHash := pow.Run()

		// Create new POW with pre-set valid nonce
		block.Header.Nonce = validNonce
		block.Hash = validHash
		pow = consensus.NewProofOfWork(block)

		newNonce, newHash := pow.Run()
		assert.Equal(t, validNonce, newNonce, "Should find existing valid nonce")
		assert.Equal(t, validHash, newHash)
	})

	t.Run("Concurrent Mining", func(t *testing.T) {
		block := createTestBlock(1)
		results := make(chan struct {
			Nonce uint64
			Hash  []byte
		}, 2)

		// Run two miners concurrently
		go func() {
			pow := consensus.NewProofOfWork(block)
			n, h := pow.Run()
			results <- struct {
				Nonce uint64
				Hash  []byte
			}{n, h}
		}()
		go func() {
			pow := consensus.NewProofOfWork(block)
			n, h := pow.Run()
			results <- struct {
				Nonce uint64
				Hash  []byte
			}{n, h}
		}()

		// Get both results
		result1 := <-results
		result2 := <-results

		// Both solutions should be valid
		assert.True(t, crypto.ValidateHash(result1.Hash, block.Header.Difficulty))
		assert.True(t, crypto.ValidateHash(result2.Hash, block.Header.Difficulty))

	})
}

func createTestBlock(index uint64) *types.Block {
	return types.NewBlock(
		index,
		[]*types.Transaction{types.NewCoinbaseTx("miner-address")},
		crypto.HashData([]byte("parent")),
		"miner-address",
	)
}
