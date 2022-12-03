package wallet

import (
	"fmt"
	"testing"

	"blockchain/foundation/cryptography"
)

func Test_Wallet(t *testing.T) {
	w, err := NewWallet(cryptography.GenerateBlockchainAddress)
	if err != nil {
		t.Errorf("Failed to instatiate a wallet with err: %s", err)
	}
	fmt.Println(w.PrivateKeyStr())
	fmt.Println(w.PublicKeyStr())
	fmt.Println(w.BlockchainAddress())

	tr := NewTransaction(w.PrivateKey(), w.PublicKey(), w.blockchainAddress, "Niko", 1.0)
	s, err := tr.GenerateSignature()
	if err != nil {
		t.Errorf("Failed to GenerateSignature  with err: %s", err)
	}

	fmt.Println(s.String())
}
