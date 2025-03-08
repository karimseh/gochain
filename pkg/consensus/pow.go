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
	b.Header.Difficulty = TargetBits
	return &ProofOfWork{Block: b}
}

func (pow *ProofOfWork) Run() (uint64, []byte) {
	var nonce uint64
	var hash []byte

	for nonce = 0; ; nonce++ {
		pow.Block.Header.Nonce = nonce
		hash = pow.Block.CalculateHash()

		if crypto.ValidateHash(hash, pow.Block.Header.Difficulty) {
			break
		}
	}
	return nonce, hash
}
