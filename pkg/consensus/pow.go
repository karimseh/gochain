package consensus

import (
	"github.com/karimseh/gochain/pkg/crypto"
	"github.com/karimseh/gochain/pkg/types"
)

const (
	TargetBits = 18
)

type ProofOfWork struct {
	Block *types.Block
}

func NewProofOfWork(b *types.Block) *ProofOfWork {

	return &ProofOfWork{
		Block: b,
	}
}

func (pow *ProofOfWork) Run() (uint64, []byte) {
	var nonce uint64
	var hash []byte

	for nonce = 0; ; nonce++ {
		hash = crypto.HashBlock(
			pow.Block.Index,
			pow.Block.Timestamp,
			pow.Block.Data,
			pow.Block.PrevHash,
			pow.Block.Difficulty,
			nonce,
		)

		if crypto.ValidateHash(hash, pow.Block.Difficulty) {
			break
		}
	}
	return nonce, hash
}
