package blockchain_server

import (
	"blockchain/blockchain-service/blockchain"
	"blockchain/blockchain-service/wallet"
	"crypto/ecdsa"
)

var cache map[string]*blockchain.Blockchain = make(map[string]*blockchain.Blockchain)

type Server struct {
	port            uint16
	generateAddress func(pKey ecdsa.PrivateKey) string
}

func New(port uint16, generateAddressFunc func(pKey ecdsa.PrivateKey) string) *Server {
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

func (s *Server) AddTransaction(sender, receiver string, amount float32) (id string, err error) {
	//TODO implement me
	panic("implement me")
}
