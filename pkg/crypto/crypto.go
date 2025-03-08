package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"math/big"
)

type PubKey *ecdsa.PublicKey
type PrivateKey *ecdsa.PrivateKey



func HashData(data ...[]byte) []byte {
	hasher := sha256.New()

	for _, d := range data {
		length := make([]byte, 8)
		binary.BigEndian.PutUint64(length, uint64(len(d)))
		hasher.Write(length)
		hasher.Write(d)
	}

	return hasher.Sum(nil)
}

func SignData(data []byte, privateKey *ecdsa.PrivateKey) ([]byte, error) {
	r, s, err := ecdsa.Sign(rand.Reader, privateKey, data)
	if err != nil {
		return nil, err
	}
	return append(r.Bytes(), s.Bytes()...), nil
}

func VerifySignature(data []byte, signature []byte, publicKey *ecdsa.PublicKey) bool {
	if len(signature) != 64 {
		return false
	}

	r := new(big.Int).SetBytes(signature[:len(signature)/2])
	s := new(big.Int).SetBytes(signature[len(signature)/2:])
	return ecdsa.Verify(publicKey, data, r, s)
}

func GenerateKeyPair() (*ecdsa.PrivateKey, error) {
	return ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
}

func NewPrivateKey(pubKeyBytes []byte, privKeyBytes []byte) *ecdsa.PrivateKey {
	curve := elliptic.P256()
	
	return &ecdsa.PrivateKey{
		PublicKey: ecdsa.PublicKey{
			Curve: curve,
			X:     new(big.Int).SetBytes(pubKeyBytes[:32]),
			Y:     new(big.Int).SetBytes(pubKeyBytes[32:]),
		},
		D: new(big.Int).SetBytes(privKeyBytes),
	}
}

func AddressFromPublicKey(pubKey *ecdsa.PublicKey) string {
	pubBytes := append(pubKey.X.Bytes(), pubKey.Y.Bytes()...)
	hash := HashData(pubBytes)
	return hex.EncodeToString(hash[:20]) // First 20 bytes
}
