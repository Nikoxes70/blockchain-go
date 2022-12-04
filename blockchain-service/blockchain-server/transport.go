package blockchain_server

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	http2 "blockchain/foundation/http"

	"blockchain/blockchain-service/blockchain"
)

type Serverer interface {
	GetBlockchain(address string) (*blockchain.Blockchain, error)
	GetTransactions() ([]*blockchain.Transaction, error)
	CreateTransaction(senderPublicKey, senderBlockchainAddress, recipientBlockchainAddress, signature string, amount float32) error
	CalculateBalance(address string) (float32, error)
}

type Transporter struct {
	server Serverer
}

func NewTransport(s Serverer) *Transporter {
	return &Transporter{
		s,
	}
}

func (t *Transporter) HandleGetChain(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		address := r.URL.Query().Get("adds")
		bc, err := t.server.GetBlockchain(address)
		if err != nil {
			http.Error(w, "404 not found.", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Add("Content-Type", "application/json")
		m, err := bc.MarshalJSON()
		if err != nil {
			http.Error(w, "blockchain-server error - failed to parse request body", http.StatusInternalServerError)
			return
		}
		io.WriteString(w, string(m[:]))
	default:
		log.Printf("ERROR: Invalid HTTP Method")
	}
}

func (t *Transporter) HandleTransactions(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		t.getTransaction(w, r)
	case http.MethodPost:
		t.createTransaction(w, r)
	default:
		http.Error(w, "page not found", http.StatusBadRequest)
	}
}

func (t *Transporter) HandleBalance(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		bcAddress := r.URL.Query().Get("bc_address")
		if bcAddress == "" {
			io.WriteString(w, "fail")
			http.Error(w, "missing blockchain address", http.StatusBadRequest)
		}
		balance, err := t.server.CalculateBalance(bcAddress)
		if err != nil {
			io.WriteString(w, "fail")
			http.Error(w, "something went wrong", http.StatusInternalServerError)
		}

		w.Header().Add("Content-Type", "application/json")
		b, err := json.Marshal(struct {
			Balance float32 `json:"balance"`
		}{
			Balance: balance,
		})
		io.WriteString(w, string(b[:]))
	case http.MethodPost:
		t.createTransaction(w, r)
	default:
		http.Error(w, "page not found", http.StatusBadRequest)
	}
}

func (t *Transporter) HandleMining(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		bc, err := t.server.GetBlockchain("BLOCKCHAIN")
		if err != nil {
			io.WriteString(w, "fail")
			http.Error(w, "blockchain not found", http.StatusNotFound)
		}
		timestamp, err := bc.Mine()
		if err != nil {
			io.WriteString(w, "fail")
			http.Error(w, "mining failed", http.StatusInternalServerError)
		}

		w.Header().Add("Content-Type", "application/json")
		b, err := json.Marshal(struct {
			Timestamp int64 `json:"timestamp"`
		}{
			timestamp,
		})
		io.WriteString(w, string(b[:]))
	default:
		http.Error(w, "page not found", http.StatusBadRequest)
	}
}

func (t *Transporter) getTransaction(w http.ResponseWriter, r *http.Request) {
	ts, err := t.server.GetTransactions()
	if err != nil {
		io.WriteString(w, "fail")
	}

	w.Header().Add("Content-Type", "application/json")
	b, err := json.Marshal(struct {
		Transactions []*blockchain.Transaction `json:"transactions"`
		Lenght       int                       `json:"lenght"`
	}{
		Transactions: ts,
		Lenght:       len(ts),
	})
	io.WriteString(w, string(b[:]))
}

func (t *Transporter) createTransaction(w http.ResponseWriter, r *http.Request) {
	var trReq blockchain.TransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&trReq); err != nil {
		http.Error(w, "blockchain-server error - failed to parse request body", http.StatusInternalServerError)
		return
	}

	if !trReq.Validate() {
		io.WriteString(w, string(http2.JsonStatus("fail")))
		http.Error(w, "bad request body", http.StatusBadRequest)
		return
	}

	err := t.server.CreateTransaction(*trReq.SenderPublicKey, *trReq.SenderBlockchainAddress, *trReq.RecipientBlockchainAddress, *trReq.Signature, *trReq.Value)
	if err != nil {
		io.WriteString(w, "failed")
		http.Error(w, "blockchain-server error - failed to create short url", http.StatusInternalServerError)
		return
	}

	io.WriteString(w, "success")
}
