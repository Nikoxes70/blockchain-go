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
	GetTransactions() ([]byte, error)
	CalculateBalance(address string) (float32, error)
	CreateTransaction(senderPublicKey, senderBlockchainAddress, recipientBlockchainAddress, signature string, amount float32) error
	AddTransaction(senderPublicKey, senderBlockchainAddress, recipientBlockchainAddress, signature string, amount float32) error
	CleaTransactionPool() int
	Mine() (int64, bool, error)
	ResolveConflicts() (bool, error)
	GetBlockchainBytes() ([]byte, error)
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
	w.Header().Add("Content-Type", "application/json")
	switch r.Method {
	case http.MethodGet:
		b, err := t.server.GetBlockchainBytes()
		if err != nil {
			io.WriteString(w, string(http2.JsonStatus("fail")))
			http.Error(w, "404 not found.", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Add("Content-Type", "application/json")
		io.WriteString(w, string(b[:]))
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
	case http.MethodPut:
		t.addTransaction(w, r)
	case http.MethodDelete:
		count := t.server.CleaTransactionPool()
		w.Header().Add("Content-Type", "application/json")
		b, err := json.Marshal(struct {
			Count int `json:"count"`
		}{
			Count: count,
		})
		if err != nil {

		}
		io.WriteString(w, string(b[:]))
	default:
		http.Error(w, "page not found", http.StatusBadRequest)
	}
}

func (t *Transporter) HandleBalance(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		bcAddress := r.URL.Query().Get("bc_address")
		if bcAddress == "" {
			io.WriteString(w, string(http2.JsonStatus("fail")))
			http.Error(w, "missing blockchain address", http.StatusBadRequest)
		}
		balance, err := t.server.CalculateBalance(bcAddress)
		if err != nil {
			io.WriteString(w, string(http2.JsonStatus("fail")))
			http.Error(w, "something went wrong", http.StatusInternalServerError)
		}

		w.Header().Add("Content-Type", "application/json")
		b, err := json.Marshal(struct {
			Balance float32 `json:"balance"`
		}{
			Balance: balance,
		})
		io.WriteString(w, string(b[:]))
	default:
		http.Error(w, "page not found", http.StatusBadRequest)
	}
}

func (t *Transporter) HandleMining(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		timestamp, mined, err := t.server.Mine()
		if err != nil {
			io.WriteString(w, string(http2.JsonStatus("fail")))
			http.Error(w, "mining failed", http.StatusInternalServerError)
		}

		w.Header().Add("Content-Type", "application/json")
		b, err := json.Marshal(struct {
			Timestamp int64 `json:"timestamp"`
			Mined     bool  `json:"mined"`
		}{
			Timestamp: timestamp,
			Mined:     mined,
		})
		io.WriteString(w, string(b[:]))
	default:
		http.Error(w, "page not found", http.StatusBadRequest)
	}
}

func (t *Transporter) HandleConsensus(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPut:
		resolved, err := t.server.ResolveConflicts()
		if err != nil {
			io.WriteString(w, string(http2.JsonStatus("fail")))
			http.Error(w, "failed to resolve conflicts", http.StatusInternalServerError)
		}

		w.Header().Add("Content-Type", "application/json")
		b, err := json.Marshal(struct {
			Resolved bool `json:"resolved"`
		}{
			Resolved: resolved,
		})
		io.WriteString(w, string(b[:]))
	default:
		http.Error(w, "page not found", http.StatusBadRequest)
	}
}

func (t *Transporter) getTransaction(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	b, err := t.server.GetTransactions()
	if err != nil {
		io.WriteString(w, string(http2.JsonStatus("fail")))
		http.Error(w, "blockchain-server error - failed to parse request body", http.StatusInternalServerError)
	}
	io.WriteString(w, string(b[:]))
}

func (t *Transporter) createTransaction(w http.ResponseWriter, r *http.Request) {
	var trReq blockchain.TransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&trReq); err != nil {
		io.WriteString(w, string(http2.JsonStatus("fail")))
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
		io.WriteString(w, string(http2.JsonStatus("fail")))
		http.Error(w, "blockchain-server error - failed to create short url", http.StatusInternalServerError)
		return
	}

	io.WriteString(w, "success")
}

func (t *Transporter) addTransaction(w http.ResponseWriter, r *http.Request) {
	var trReq blockchain.TransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&trReq); err != nil {
		io.WriteString(w, string(http2.JsonStatus("fail")))
		http.Error(w, "blockchain-server error - failed to parse request body", http.StatusInternalServerError)
		return
	}

	if !trReq.Validate() {
		io.WriteString(w, string(http2.JsonStatus("fail")))
		http.Error(w, "bad request body", http.StatusBadRequest)
		return
	}

	err := t.server.AddTransaction(*trReq.SenderPublicKey, *trReq.SenderBlockchainAddress, *trReq.RecipientBlockchainAddress, *trReq.Signature, *trReq.Value)
	if err != nil {
		io.WriteString(w, string(http2.JsonStatus("fail")))
		http.Error(w, "blockchain-server error - failed to create short url", http.StatusInternalServerError)
		return
	}

	io.WriteString(w, "success")
}
