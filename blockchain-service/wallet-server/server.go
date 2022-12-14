package wallet_server

import (
	"blockchain/blockchain-service/blockchain"
	"blockchain/blockchain-service/wallet"
	"blockchain/foundation/cryptography"
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"path"
	"strconv"
)

const (
	temoDir   = "blockchain-service/wallet-server/templates"
	indexFile = "index.html"
)

type Walleter interface {
}

type Server struct {
	port          uint16
	gateway       string
	walletService Walleter
}

func New(port uint16, gateway string) *Server {
	return &Server{
		port:    port,
		gateway: gateway,
	}
}

func (s *Server) Port() uint16 {
	return s.port
}

func (s *Server) Gateway() string {
	return s.gateway
}

func (s *Server) Index() (*template.Template, error) {
	return template.ParseFiles(path.Join(temoDir, indexFile))
}

func (s *Server) Balance(bcAddress string) ([]byte, error) {
	url := s.gateway + "/balance"
	client := http.Client{}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	q := req.URL.Query()
	q.Add("bc_address", bcAddress)
	req.URL.RawQuery = q.Encode()
	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}
	if err != nil || resp.StatusCode > 400 {
		return nil, fmt.Errorf("failed to POST url - %v, err: %s", url, err)
	}
	decoder := json.NewDecoder(resp.Body)
	var balanceResponse struct {
		Balance float32 `json:"balance"`
	}

	if err = decoder.Decode(&balanceResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response - %v, to struct: %v", resp.Body, balanceResponse)
	}

	b, err := json.Marshal(balanceResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal balance - %v, with err: %s", balanceResponse, err)
	}

	return b, nil
}

func (s *Server) Wallet() ([]byte, error) {
	w, err := wallet.NewWallet(cryptography.GenerateBlockchainAddress)
	if err != nil {
		return nil, err
	}
	b, err := json.Marshal(w)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (s *Server) CreateTransaction(senderPublicKey, senderPrivateKey, senderBlockchainAddress, recipientBlockchainAddress, v *string) ([]byte, error) {
	publicKey, err := cryptography.PublicKeyFromString(*senderPublicKey)
	if err != nil {
		return nil, err
	}

	privateKey, err := cryptography.PrivateKeyFromString(*senderPrivateKey, publicKey)
	if err != nil {
		return nil, err
	}

	value, err := strconv.ParseFloat(*v, 32)
	if err != nil {
		return nil, err
	}
	value32 := float32(value)

	walletTransaction := wallet.NewTransaction(privateKey, publicKey, *senderBlockchainAddress, *recipientBlockchainAddress, float32(value))
	sign, err := walletTransaction.GenerateSignature()
	if err != nil {
		return nil, err
	}
	sString := sign.String()

	btr := blockchain.TransactionRequest{
		SenderBlockchainAddress:    senderBlockchainAddress,
		RecipientBlockchainAddress: recipientBlockchainAddress,
		SenderPublicKey:            senderPublicKey,
		Value:                      &value32,
		Signature:                  &sString,
	}

	b, err := json.Marshal(btr)
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(b)
	url := s.gateway + "/transactions"
	resp, err := http.Post(url, "application/json", buf)
	if err != nil {
		return nil, err
	}
	if err != nil || resp.StatusCode > 400 {
		return nil, fmt.Errorf("failed to POST url - %v, err: %v", url, err)
	}

	return nil, nil
}
