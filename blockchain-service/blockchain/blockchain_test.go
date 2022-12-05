package blockchain

import (
	"blockchain/blockchain-service/wallet"
	"blockchain/foundation/cryptography"
	"fmt"
	"testing"
)

func Test_Blockchain(t *testing.T) {
	miner, err := wallet.NewWallet(cryptography.GenerateBlockchainAddress)
	if err != nil {
		t.Errorf("Failed to instatiate a wallet with err: %s", err)
	}

	niko, err := wallet.NewWallet(cryptography.GenerateBlockchainAddress)
	if err != nil {
		t.Errorf("Failed to instatiate a wallet with err: %s", err)
	}

	itay, err := wallet.NewWallet(cryptography.GenerateBlockchainAddress)
	if err != nil {
		t.Errorf("Failed to instatiate a wallet with err: %s", err)
	}

	bc, err := NewBlockchain(miner.BlockchainAddress())
	if err != nil {
		t.Errorf("Failed to instatiate a blockchain with err: %s", err)
	}
	tr := wallet.NewTransaction(itay.PrivateKey(), itay.PublicKey(), itay.BlockchainAddress(), niko.BlockchainAddress(), 1.0)

	s, err := tr.GenerateSignature()
	if err != nil {
		t.Errorf("Failed to GenerateSignature with err: %s", err)
	}

	err = bc.AddTransaction(itay.BlockchainAddress(), niko.BlockchainAddress(), 1.0, itay.PublicKey(), s)
	if err != nil {
		t.Errorf("Failed to CreateTransaction with err: %s", err)
	}

	block, mined, err := bc.Mine()
	if err != nil {
		t.Errorf("Failed to Mine with err: %s", err)
	}
	if !mined {
		t.Errorf("BLOCK NOT MINED")
	}
	fmt.Println(block)

	balance := bc.CalculateBalance(niko.BlockchainAddress())
	if balance != 1 {
		t.Errorf("Wrong calculation")
	}
	//
	//bc.CreateTransaction("Niko", "Itay", 1)
	//bc.CreateTransaction("Niko", "Itay", 2)
	//bc.CreateTransaction("Niko", "Itay", 3)
	//bc.CreateTransaction("Niko", "Itay", 4)
	//bc.Mine()
	//balance = bc.CalculateBalance("Niko")
	//if balance != 10 {
	//	t.Errorf("Wrong calculation")
	//}
	//
	//bc.CreateTransaction("Niko", "Itay", 1)
	//bc.CreateTransaction("Itay", "Niko", 4)
	//bc.CreateTransaction("Niko", "Itay", 5)
	//bc.CreateTransaction("Niko", "Itay", 1)
	//bc.Mine()
	//balance = bc.CalculateBalance("Niko")
	//if balance != 7 {
	//	t.Errorf("Wrong calculation")
	//}
	//
	//bc.CreateTransaction("Itay", "Niko", 10)
	//bc.Mine()
	//
	//balance = bc.CalculateBalance("Niko")
	//if balance != 17 {
	//	t.Errorf("Wrong calculation")
	//}
	bc.Print()
}
