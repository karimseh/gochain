package crypto

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math/big"
)

func HashBlock(index uint64, timestamp int64, data []byte, prevHash []byte, difficulty int, nonce uint64) []byte {
	headers := bytes.Join(
		[][]byte{
			toBytes(index),
			toBytes(timestamp),
			data,
			prevHash,
			toBytes(int64(difficulty)),
			toBytes(nonce),
		},
		[]byte{},
	)
	hash := sha256.Sum256(headers)
	return hash[:]
}

func toBytes(data interface{}) []byte {
	buff := new(bytes.Buffer)
	
	switch d := data.(type) {
	case uint64:
		_ = binary.Write(buff, binary.BigEndian, d)
	case int64:
		_ = binary.Write(buff, binary.BigEndian, d)
	case int:
		_ = binary.Write(buff, binary.BigEndian, int64(d))
	case []byte:
		return d
	default:
		panic(fmt.Sprintf("unsupported type: %T", data))
	}
	
	return buff.Bytes()
}

func ValidateHash(hash []byte, difficulty int) bool {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-difficulty))
	hashInt := new(big.Int).SetBytes(hash)
	return hashInt.Cmp(target) == -1
}