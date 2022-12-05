package wallet

import (
	"blockchain/foundation/cryptography"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/json"
)

type Wallet struct {
	privateKey        *ecdsa.PrivateKey
	publicKey         *ecdsa.PublicKey
	blockchainAddress string
}

func NewWallet(generateAddress func(pKey *ecdsa.PublicKey) string) (*Wallet, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}

	address := generateAddress(&privateKey.PublicKey)

	return &Wallet{
		privateKey:        privateKey,
		publicKey:         &privateKey.PublicKey,
		blockchainAddress: address,
	}, nil
}

func (w *Wallet) PrivateKey() *ecdsa.PrivateKey {
	return w.privateKey
}

func (w *Wallet) PrivateKeyStr() string {
	return cryptography.GeneratePrivateKeyString(w.privateKey.D.Bytes())
}

func (w *Wallet) PublicKey() *ecdsa.PublicKey {
	return w.publicKey
}

func (w *Wallet) PublicKeyStr() string {
	return cryptography.GeneratePublicKeyString(w.privateKey.PublicKey.X.Bytes(), w.privateKey.PublicKey.Y.Bytes())
}

func (w *Wallet) BlockchainAddress() string {
	return w.blockchainAddress
}

func (w *Wallet) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		PrivateKey        string `json:"private_key"`
		PublicKey         string `json:"public_key"`
		BlockchainAddress string `json:"blockchain_address"`
	}{
		PrivateKey:        w.PrivateKeyStr(),
		PublicKey:         w.PublicKeyStr(),
		BlockchainAddress: w.BlockchainAddress(),
	})
}
