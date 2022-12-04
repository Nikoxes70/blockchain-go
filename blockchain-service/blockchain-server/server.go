package blockchain_server

import (
	"blockchain/blockchain-service/blockchain"
	"blockchain/blockchain-service/wallet"
	"blockchain/foundation/cryptography"
	"crypto/ecdsa"
)

var cache map[string]*blockchain.Blockchain = make(map[string]*blockchain.Blockchain)

type Server struct {
	port            uint16
	generateAddress func(pKey *ecdsa.PublicKey) string
}

func New(port uint16, generateAddressFunc func(pKey *ecdsa.PublicKey) string) *Server {
	return &Server{
		port:            port,
		generateAddress: generateAddressFunc,
	}
}

func (s *Server) Port() uint16 {
	return s.port
}

func (s *Server) GetBlockchain(address string) (*blockchain.Blockchain, error) {
	bc, ok := cache[address]
	if !ok {
		minersWallet, err := wallet.NewWallet(s.generateAddress)
		if err != nil {
			return nil, err
		}
		bc, err = blockchain.NewBlockchain(minersWallet.BlockchainAddress(), s.Port())
		cache[address] = bc
	}
	return bc, nil
}

func (s *Server) GetTransactions() ([]*blockchain.Transaction, error) {
	bc, err := s.GetBlockchain("BLOCKCHAIN")
	if err != nil {
		return nil, err
	}
	return bc.TransactonPool(), err
}

func (s *Server) CreateTransaction(senderPublicKey, senderBlockchainAddress, recipientBlockchainAddress, signature string, amount float32) error {
	publicKey, err := cryptography.PublicKeyFromString(senderPublicKey)
	if err != nil {
		return err
	}

	sign, err := cryptography.SignatureFromString(signature)
	bc, err := s.GetBlockchain("BLOCKCHAIN")
	if err != nil {
		return err
	}

	if err = bc.CreateTransaction(senderBlockchainAddress, recipientBlockchainAddress, amount, publicKey, sign); err != nil {
		return err
	}

	return nil
}
