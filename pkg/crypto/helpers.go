package crypto

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"fmt"

	"encoding/binary"
	"encoding/gob"

	"math/big"
)

func Uint64ToBytes(num uint64) []byte {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, num)
	return buf
}

func Serialize(data interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(data)
	return buf.Bytes(), err
}

func ValidateHash(hash []byte, difficulty int) bool {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-difficulty))
	hashInt := new(big.Int).SetBytes(hash)
	return hashInt.Cmp(target) == -1
}

func BuildMerkleRoot(hashes [][]byte) []byte {
	if len(hashes) == 0 {
		return nil
	}
	level := make([][]byte, len(hashes))
	copy(level, hashes)
	for len(level) > 1 {
		var nextLevel [][]byte

		for i := 0; i < len(level); i += 2 {
			var left, right []byte
			left = level[i]

			if i+1 < len(level) {
				right = level[i+1]
			} else {
				right = left
			}

			combined := HashData(left, right)
			nextLevel = append(nextLevel, combined)
		}
		level = nextLevel
	}
	return level[0]
}

// Verify if tx is included
func VerifyMerkleRoot(leaf []byte, proof [][]byte, root []byte) bool {
	current := leaf
	for _, p := range proof {
		if bytes.Compare(current, p) < 0 {
			current = HashData(current, p)
		} else {
			current = HashData(p, current)
		}
	}
	return bytes.Equal(current, root)
}

func PublicKeyToBytes(pub *ecdsa.PublicKey) []byte {
    return append(pub.X.Bytes(), pub.Y.Bytes()...)
}

func BytesToPublicKey(data []byte) (*ecdsa.PublicKey, error) {
    if len(data) != 64 {
        return nil, fmt.Errorf("invalid public key length")
    }
    
    curve := elliptic.P256()
    x := new(big.Int).SetBytes(data[:32])
    y := new(big.Int).SetBytes(data[32:64])
    
    if !curve.IsOnCurve(x, y) {
        return nil, fmt.Errorf("invalid public key coordinates")
    }
    
    return &ecdsa.PublicKey{
        Curve: curve,
        X:     x,
        Y:     y,
    }, nil
}
