package types

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/karimseh/gochain/pkg/crypto"
)

type Block struct {
	mu           sync.RWMutex
	Header       BlockHeader    `json:"header"`
	Transactions []*Transaction `json:"transactions"`
	MerkleRoot   []byte         `json:"merkle_root"`
	StateRoot    []byte         `json:"state_root"`
	Hash         []byte         `json:"hash"`
}
type BlockHeader struct {
	ParentHash []byte `json:"parentHash"`
	Index      uint64 `json:"index"`
	Timestamp  int64  `json:"timestamp"`
	Nonce      uint64 `json:"nonce"`
	Difficulty int    `json:"difficulty"`
	Miner      string `json:"miner"`
}

func NewBlock(index uint64, transactions []*Transaction, parentHash []byte, miner string) *Block {
	header := BlockHeader{
		Index:      index,
		Timestamp:  time.Now().Unix(),
		ParentHash: parentHash,
		Miner:      miner,
	}

	block := &Block{
		Header:       header,
		Transactions: transactions,
		MerkleRoot:   CalculateMerkleRoot(transactions),
	}
	block.Hash = block.CalculateHash()
	return block
}

func (b *Block) SetNonce(nonce uint64) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.Header.Nonce = nonce
}

func (b *Block) GetDifficulty() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.Header.Difficulty
}
func (b *Block) SetDifficulty(diff int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.Header.Difficulty = diff
}

func (b *Block) Validate() error {
	b.mu.RLock()
	defer b.mu.RUnlock()
	// Temporary clear existing hash for validation
	storedHash := b.Hash
	b.Hash = nil

	calculatedHash := b.CalculateHash()

	b.Hash = storedHash // Restore original hash

	if !bytes.Equal(calculatedHash, storedHash) {
		return fmt.Errorf("invalid block hash\nExpected: %x\nActual:   %x",
			calculatedHash,
			storedHash,
		)
	}

	if !crypto.ValidateHash(storedHash, b.Header.Difficulty) {
		return fmt.Errorf("hash doesn't meet difficulty requirements")
	}
	if b.Header.Miner == "" {
		return fmt.Errorf("invalid miner address")
	}
	if b.Header.Timestamp == 0 {
		return fmt.Errorf("invalid timestamp")
	}
	if b.Header.Index == 0 {
		return fmt.Errorf("invalid block index")
	}
	if b.Header.ParentHash == nil {
		return fmt.Errorf("invalid parent hash")
	}
	if b.MerkleRoot == nil {
		return fmt.Errorf("invalid merkle root")
	}

	return nil
}

func (b *Block) CalculateHash() []byte {
	b.mu.RLock()
	defer b.mu.RUnlock()
	headerData, _ := crypto.Serialize(b.Header)
	return crypto.HashData(headerData, b.MerkleRoot, b.StateRoot)
}

func CalculateMerkleRoot(txs []*Transaction) []byte {
	var hashes [][]byte
	for _, tx := range txs {
		hashes = append(hashes, tx.Hash)
	}
	return crypto.BuildMerkleRoot(hashes)
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
