package consensus

import (
	"github.com/karimseh/gochain/pkg/crypto"
	"github.com/karimseh/gochain/pkg/types"
)

const (
	TargetBits = 18
)

type ProofOfWork struct {
	block *types.Block
}

func NewProofOfWork(b *types.Block) *ProofOfWork {
	b.SetDifficulty(TargetBits)
	return &ProofOfWork{block: b}
}

func (pow *ProofOfWork) Run() (uint64, []byte) {
	var nonce uint64
	var hash []byte

	for nonce = 0; ; nonce++ {
		pow.block.SetNonce(nonce)
		hash = pow.block.CalculateHash()

		if crypto.ValidateHash(hash, pow.block.GetDifficulty()) {
			break
		}
	}
	return nonce, hash
}
