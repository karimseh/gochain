package types

import (
	"bytes"

	"github.com/karimseh/gochain/pkg/crypto"
	"github.com/karimseh/gochain/pkg/wallet"
)

const (
	CoinbaseAmount = 50 // Block reward
	CoinbaseNonce  = 0
)

type Transaction struct {
	From      string `json:"from"`
	To        string `json:"to"`
	Ammount   uint64 `json:"ammount"`
	Nonce     uint64 `json:"nonce"`
	Signature []byte `json:"signature"`
	Hash      []byte `json:"hash"`
	PubKey    []byte `json:"pubkey"`
}

func NewTransaction(from, to string, ammount, nonce uint64, pubKey []byte) *Transaction {
	return &Transaction{
		From:    from,
		To:      to,
		Ammount: ammount,
		Nonce:   nonce,
		PubKey:  pubKey,
	}
}

func (tx *Transaction) Sign(wallet *wallet.Wallet) error {
	tx.Hash = tx.CalculateHash()
	signature, err := crypto.SignData(tx.Hash, wallet.PrivateKey)
	if err != nil {
		return err
	}
	tx.Signature = signature
	return nil
}

func (tx *Transaction) Verify() bool {
	if tx.IsCoinbase() {
		return tx.Ammount == CoinbaseAmount &&
			tx.To != "" &&
			len(tx.Signature) == 0
	}
	if !bytes.Equal(tx.Hash, tx.CalculateHash()) {
		return false
	}
	pubKey, err := crypto.BytesToPublicKey(tx.PubKey)
	if err != nil {
		return false
	}
	if len(tx.Signature) == 0 {
		return false
	}
	if !crypto.VerifySignature(tx.Hash, tx.Signature, pubKey) {
		return false
	}
	return crypto.AddressFromPublicKey(pubKey) == tx.From
}

func (tx *Transaction) CalculateHash() []byte {
	return crypto.HashData([]byte(tx.From), []byte(tx.To), crypto.Uint64ToBytes(tx.Ammount), crypto.Uint64ToBytes(tx.Nonce))
}

func NewCoinbaseTx(minerAddress string) *Transaction {
	tx := &Transaction{
		To:      minerAddress,
		Ammount: CoinbaseAmount,
		Nonce:   CoinbaseNonce,
	}
	tx.Hash = crypto.HashData()
	return tx
}

func (tx *Transaction) IsCoinbase() bool {
	return tx.Nonce == CoinbaseNonce && tx.From == ""
}
