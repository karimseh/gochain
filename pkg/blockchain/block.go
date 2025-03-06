package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
)

type Block struct {
	Index     uint64 `json:"index"`
	Timestamp int64  `json:"timestamp"`
	Data      []byte `json:"data"`
	PrevHash  []byte `json:"prevHash"`
	Hash      []byte `json:"hash"`
	Nonce     uint64 `json:"nonce"`
}

func (b *Block) Serialize() []byte {
	data, _ := json.Marshal(b)
	return data
}

func DeserializeBlock(data []byte) (*Block, error) {
	var block Block
	if err := json.Unmarshal(data, &block); err != nil {
		return nil, err
	}
	return &block, nil
}

func (b *Block) Validate() error {
	if len(b.Hash) == 0 {
		return fmt.Errorf("block hash is empty")
	}
	calculatedHash := b.CalculateHash()
	if !bytes.Equal(calculatedHash, b.Hash) {
		return fmt.Errorf("invalid block hash")
	}
	return nil
}

func (b *Block) CalculateHash() []byte {
	headers := fmt.Sprintf("%d%d%x%x%d", b.Index, b.Timestamp, b.Data, b.PrevHash, b.Nonce)

	hash := sha256.Sum256([]byte(headers))
	return hash[:]
}
