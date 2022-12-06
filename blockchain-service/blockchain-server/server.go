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
	"blockchain/foundation/cryptography"
	"blockchain/foundation/network"
)

const (
	BLOCKCHAIN_PORT_RANGE_START = 5000
	BLOCKCHAIN_PORT_RANGE_END   = 5003
	NEIGHBOR_IP_RANGE_START     = 0
	NEIGHBOR_IP_RANGE_END       = 1
)

type blockchainer interface {
	SetChain(c []*blockchain.Block)
	Chain() []*blockchain.Block
	TransactionPool() []*blockchain.Transaction
	TruncateTransactionPool() int
	CreateTransaction(sender, recipient string, value float32, pKey *ecdsa.PublicKey, s *cryptography.Signature) error
	AddTransaction(sender, recipient string, value float32, pKey *ecdsa.PublicKey, s *cryptography.Signature) error
	Mine() (int64, bool, error)
	CalculateBalance(address string) float32
	ValidChain(chain []*blockchain.Block) (bool, error)
	Print()
	MarshalJSON() ([]byte, error)
}

type Server struct {
	port uint16
	mux  sync.Mutex
	bc   blockchainer

	neighbors       []string
	muxNeighbors    sync.Mutex
	generateAddress func(pKey *ecdsa.PublicKey) string
}

func New(port uint16, bc blockchainer, generateAddressFunc func(pKey *ecdsa.PublicKey) string) *Server {
	return &Server{
		port:            port,
		bc:              bc,
		muxNeighbors:    sync.Mutex{},
		generateAddress: generateAddressFunc,
	}
}

func (s *Server) GetBlockchainBytes() ([]byte, error) {
	return s.bc.MarshalJSON()
}

func (s *Server) Mine() (int64, bool, error) {
	t, mined, err := s.bc.Mine()
	if err != nil {
		return 0, false, err
	}

	if mined {
		updatedCount, err := s.DeleteNeighborsPools()
		if err != nil {
			log.Printf("failed to DeleteNeighborsPools with err: %s", err)
		}
		log.Printf("updated %d nneighbors", updatedCount)

		var errsStr []string
		for _, n := range s.neighbors {
			endpoint := fmt.Sprintf("http://%s/consensus", n)
			client := &http.Client{}
			req, err := http.NewRequest(http.MethodPut, endpoint, nil)
			if err != nil {
				errsStr = append(errsStr, err.Error())
				continue
			}
			resp, err := client.Do(req)
			if err != nil {
				errsStr = append(errsStr, err.Error())
				continue
			}
			if resp.StatusCode > 400 {
				errsStr = append(errsStr, fmt.Sprintf("Request failed - status: %s, body: %v", resp.Status, resp.Body))
				continue
			}
		}
	}

	return t, mined, nil
}

func (s *Server) CalculateBalance(address string) (float32, error) {
	return s.bc.CalculateBalance(address), nil
}

func (s *Server) GetTransactions() []*blockchain.Transaction {
	return s.bc.TransactionPool()
}

func (s *Server) CreateTransaction(senderPublicKey, senderBlockchainAddress, recipientBlockchainAddress, signature string, amount float32) error {
	publicKey, err := cryptography.PublicKeyFromString(senderPublicKey)
	if err != nil {
		return err
	}

	sign, err := cryptography.SignatureFromString(signature)

	if err = s.bc.CreateTransaction(senderBlockchainAddress, recipientBlockchainAddress, amount, publicKey, sign); err != nil {
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

	if err = s.bc.AddTransaction(senderBlockchainAddress, recipientBlockchainAddress, amount, publicKey, sign); err != nil {
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

	if errsStr != nil {
		return neightborsUpdated, fmt.Errorf(strings.Join(errsStr, "\n"))
	}
	return neightborsUpdated, nil
}

func (s *Server) DeleteNeighborsPools() (int, error) {
	neightborsUpdated := 0
	var errsStr []string
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

func (s *Server) CleaTransactionPool() int {
	return s.bc.TruncateTransactionPool()
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

func (s *Server) ResolveConflicts() (bool, error) {
	var longestChain []*blockchain.Block
	maxLenght := len(s.bc.Chain())
	var errsStr []string
	for _, n := range s.neighbors {
		endpoint := fmt.Sprintf("http://%s/chain", n)
		resp, err := http.Get(endpoint)
		if err != nil {
			errsStr = append(errsStr, err.Error())
			continue
		}

		if err != nil || resp.StatusCode > 400 {
			errsStr = append(errsStr, err.Error())
			continue
		}
		decoder := json.NewDecoder(resp.Body)
		var otherBlockchain blockchain.Blockchain
		err = decoder.Decode(&otherBlockchain)
		if err != nil {
			errsStr = append(errsStr, err.Error())
			continue
		}
		chain := otherBlockchain.Chain()
		if len(chain) > maxLenght {
			valid, err := s.bc.ValidChain(chain)
			if err != nil {
				errsStr = append(errsStr, err.Error())
				continue
			}

			if valid {
				maxLenght = len(chain)
				longestChain = chain
			}
		}
	}

	if errsStr != nil {
		return false, fmt.Errorf(strings.Join(errsStr, "\n"))
	}

	if longestChain != nil {
		s.bc.SetChain(longestChain)
		return true, nil
	}
	return false, nil
}
