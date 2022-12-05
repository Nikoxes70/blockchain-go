package blockchain_server

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"

	"blockchain/blockchain-service/blockchain"
	"blockchain/blockchain-service/wallet"
	"blockchain/foundation/cryptography"
	"blockchain/foundation/network"
)

const (
	BLOCKCHAIN_PORT_RANGE_START = 5000
	BLOCKCHAIN_PORT_RANGE_END   = 5003
	NEIGHBOR_IP_RANGE_START     = 0
	NEIGHBOR_IP_RANGE_END       = 1
)

var cache map[string]*blockchain.Blockchain = make(map[string]*blockchain.Blockchain)

type Server struct {
	port uint16
	mux  sync.Mutex

	neighbors       []string
	muxNeighbors    sync.Mutex
	generateAddress func(pKey *ecdsa.PublicKey) string
}

func New(port uint16, generateAddressFunc func(pKey *ecdsa.PublicKey) string) *Server {
	return &Server{
		port:            port,
		generateAddress: generateAddressFunc,
	}
}

func (s *Server) Mine() (int64, bool, error) {
	bc, err := s.GetBlockchain("BLOCKCHAIN")
	if err != nil {
		return 0, false, err
	}
	t, mined, err := bc.Mine()
	if err != nil {
		return 0, false, err
	}

	updatedCount, err := s.DeleteNeighborsPools()
	if err != nil {
		log.Printf("failed to DeleteNeighborsPools with err: %s", err)
	}
	log.Printf("updated %d nneighbors", updatedCount)
	return t, mined, nil
}

func (s *Server) GetBlockchain(address string) (*blockchain.Blockchain, error) {
	bc, ok := cache[address]
	if !ok {
		minersWallet, err := wallet.NewWallet(s.generateAddress)
		if err != nil {
			return nil, err
		}
		bc, err = blockchain.NewBlockchain(minersWallet.BlockchainAddress())
		cache[address] = bc
	}
	return bc, nil
}

func (s *Server) CalculateBalance(address string) (float32, error) {
	bc, err := s.GetBlockchain("BLOCKCHAIN")
	if err != nil {
		return 0, err
	}
	return bc.CalculateBalance(address), nil
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

func (s *Server) AddTransaction(senderPublicKey, senderBlockchainAddress, recipientBlockchainAddress, signature string, amount float32) error {
	publicKey, err := cryptography.PublicKeyFromString(senderPublicKey)
	if err != nil {
		return err
	}

	sign, err := cryptography.SignatureFromString(signature)
	bc, err := s.GetBlockchain("BLOCKCHAIN")
	if err != nil {
		return err
	}

	if err = bc.AddTransaction(senderBlockchainAddress, recipientBlockchainAddress, amount, publicKey, sign); err != nil {
		return err
	}

	return nil
}

func (s *Server) UpdateNeighbors(senderPublicKey, senderBlockchainAddress, recipientBlockchainAddress, signature *string, amount *float32) (int, error) {
	btr := blockchain.TransactionRequest{
		SenderBlockchainAddress:    senderBlockchainAddress,
		RecipientBlockchainAddress: recipientBlockchainAddress,
		SenderPublicKey:            senderPublicKey,
		Value:                      amount,
		Signature:                  signature,
	}

	b, err := json.Marshal(btr)
	if err != nil {
		return 0, err
	}

	neightborsUpdated := 0
	errsStr := []string{}
	for _, n := range s.neighbors {
		buf := bytes.NewBuffer(b)
		endpoint := fmt.Sprintf("http://%s/transactions", n)
		client := &http.Client{}
		req, err := http.NewRequest(http.MethodPut, endpoint, buf)
		if err != nil {
			errsStr = append(errsStr, err.Error())
			continue
		}
		resp, err := client.Do(req)
		if err != nil {
			errsStr = append(errsStr, err.Error())
			continue
		}
		neightborsUpdated++
		log.Printf("%v", resp)
	}

	return neightborsUpdated, fmt.Errorf(strings.Join(errsStr, "\n"))
}

func (s *Server) DeleteNeighborsPools() (int, error) {
	neightborsUpdated := 0
	errsStr := []string{}
	for _, n := range s.neighbors {
		endpoint := fmt.Sprintf("http://%s/transactions", n)
		client := &http.Client{}
		req, err := http.NewRequest(http.MethodDelete, endpoint, nil)
		if err != nil {
			errsStr = append(errsStr, err.Error())
			continue
		}
		resp, err := client.Do(req)
		if err != nil {
			errsStr = append(errsStr, err.Error())
			continue
		}
		neightborsUpdated++
		log.Printf("%v", resp)
	}

	return neightborsUpdated, fmt.Errorf(strings.Join(errsStr, "\n"))
}

func (s *Server) CleaTransactionPool() (int, error) {
	bc, err := s.GetBlockchain("BLOCKCHAIN")
	if err != nil {
		return 0, err
	}
	c := bc.TruncateTransactonPool()
	return c, err
}

func (s *Server) SetNeighbors() (int, error) {
	n, err := network.FindNeighbors(
		network.GetHost(),
		s.port,
		NEIGHBOR_IP_RANGE_START,
		NEIGHBOR_IP_RANGE_END,
		BLOCKCHAIN_PORT_RANGE_START,
		BLOCKCHAIN_PORT_RANGE_END,
	)
	if err != nil {
		return 0, err
	}
	s.neighbors = n
	return len(n), nil
}

func (s *Server) SyncNeighbors() (int, error) {
	s.muxNeighbors.Lock()
	defer s.muxNeighbors.Unlock()
	return s.SetNeighbors()
}
