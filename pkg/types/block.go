package types

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"

	"github.com/karimseh/gochain/pkg/crypto"
)

type Block struct {
	Index      uint64 `json:"index"`
	Timestamp  int64  `json:"timestamp"`
	Data       []byte `json:"data"`
	PrevHash   []byte `json:"prevHash"`
	Hash       []byte `json:"hash"`
	Nonce      uint64 `json:"nonce"`
	Difficulty int    `json:"difficulty"`
}

func (b *Block) Serialize() []byte {
	data, _ := json.Marshal(b)
	return data
}

func DeserializeBlock(data []byte) (*Block, error) {
	var block Block
	err := json.Unmarshal(data, &block)
	return &block, err
}

func (b *Block) Validate() error {
	// Temporary clear existing hash for validation
	storedHash := b.Hash
	b.Hash = nil

	calculatedHash := crypto.HashBlock(
		b.Index,
		b.Timestamp,
		b.Data,
		b.PrevHash,
		b.Difficulty,
		b.Nonce,
	)

	b.Hash = storedHash // Restore original hash

	if !bytes.Equal(calculatedHash, storedHash) {
		return fmt.Errorf("invalid block hash\nExpected: %x\nActual:   %x",
			calculatedHash,
			storedHash,
		)
	}

	if !crypto.ValidateHash(storedHash, b.Difficulty) {
		return fmt.Errorf("hash doesn't meet difficulty requirements")
	}

	return nil
}

func NewBlock(index uint64, data []byte, prevHash []byte, difficulty int) *Block {
	return &Block{
		Index:      index,
		Timestamp:  time.Now().Unix(),
		Data:       data,
		PrevHash:   prevHash,
		Difficulty: difficulty,
		// Hash and Nonce will be set during mining
	}
}

func (b *Block) CalculateHash() []byte {
	headers := fmt.Sprintf("%d%d%x%x%d%d",
		b.Index,
		b.Timestamp,
		b.Data,
		b.PrevHash,
		b.Nonce,
		b.Difficulty,
	)
	hash := sha256.Sum256([]byte(headers))
	return hash[:]
}
