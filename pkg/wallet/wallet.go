package wallet

import (
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/karimseh/gochain/pkg/crypto"
)

const walletDir = "wallets"

type Wallet struct {
	PrivateKey crypto.PrivateKey `json:"private_key"`
	PublicKey  crypto.PubKey     `json:"public_key"`
	Address    string            `json:"address"`
}

func NewWallet() *Wallet {
	privateKey, err := crypto.GenerateKeyPair()
	if err != nil {
		panic("failed to generate key pair: " + err.Error())
	}

	publicKey := &privateKey.PublicKey
	return &Wallet{
		PrivateKey: privateKey,
		PublicKey:  publicKey,
		Address:    crypto.AddressFromPublicKey(publicKey),
	}
}

func (w *Wallet) SaveToFile() error {
	if err := os.MkdirAll(walletDir, 0700); err != nil {
		return err
	}

	walletFile := filepath.Join(walletDir, w.Address+".json")
	data := struct {
		PrivateKey string `json:"privateKey"`
		PublicKey  string `json:"publicKey"`
	}{
		PrivateKey: hex.EncodeToString(w.PrivateKey.D.Bytes()),
		PublicKey:  hex.EncodeToString(append(w.PublicKey.X.Bytes(), w.PublicKey.Y.Bytes()...)),
	}

	file, err := json.MarshalIndent(data, "", " ")
	if err != nil {
		return err
	}

	return os.WriteFile(walletFile, file, 0600)
}

func LoadWallet(address string) (*Wallet, error) {
	walletFile := filepath.Join(walletDir, address+".json")
	data, err := os.ReadFile(walletFile)
	if err != nil {
		return nil, err
	}

	var walletData struct {
		PrivateKey string `json:"privateKey"`
		PublicKey  string `json:"publicKey"`
	}
	if err := json.Unmarshal(data, &walletData); err != nil {
		return nil, err
	}

	privKeyBytes, err := hex.DecodeString(walletData.PrivateKey)
	if err != nil {
		return nil, err
	}

	pubKeyBytes, err := hex.DecodeString(walletData.PublicKey)
	if err != nil {
		return nil, err
	}

	privKey := crypto.NewPrivateKey(pubKeyBytes, privKeyBytes)

	return &Wallet{
		PrivateKey: privKey,
		PublicKey:  &privKey.PublicKey,
		Address:    address,
	}, nil
}
