package wallet_server

import (
	http2 "blockchain/foundation/http"
	"encoding/json"
	"html/template"
	"io"
	"net/http"
)

type TransactionRequest struct {
	SenderPrivateKey           *string `json:"sender_private_key"`
	SenderPublicKey            *string `json:"sender_public_key"`
	SenderBlockchainAddress    *string `json:"sender_blockchain_address"`
	RecipientBlockchainAddress *string `json:"recipient_blockchain_address"`
	Value                      *string `json:"value"`
}

func (tr *TransactionRequest) Validate() bool {
	return tr.RecipientBlockchainAddress != nil &&
		tr.SenderBlockchainAddress != nil &&
		tr.Value != nil && tr.SenderPrivateKey != nil &&
		tr.SenderPublicKey != nil
}

type Serverer interface {
	Index() (*template.Template, error)
	Wallet() ([]byte, error)
	CreateTransaction(senderPublicKey, senderPrivateKey, senderBlockchainAddress, recipientBlockchainAddress, v string) ([]byte, error)
}

type Transporter struct {
	server Serverer
}

func NewTransport(s Serverer) *Transporter {
	return &Transporter{
		s,
	}
}

func (t *Transporter) HandleIndex(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		t, err := t.server.Index()
		if err != nil {
			http.Error(w, "server error - page not found", http.StatusInternalServerError)
			return
		}
		if err := t.Execute(w, ""); err != nil {
			http.Error(w, "server error - page execution failed ", http.StatusInternalServerError)
			return
		}
	default:
		http.Error(w, "page not found", http.StatusBadRequest)
	}
}

func (t *Transporter) HandleWallet(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		b, err := t.server.Wallet()
		if err != nil {
			http.Error(w, "server error - page not found", http.StatusInternalServerError)
			return
		}
		w.Header().Add("Content-Type", "application/json")
		io.WriteString(w, string(b[:]))
	default:
		http.Error(w, "page not found", http.StatusBadRequest)
	}
}

func (t *Transporter) HandleTransaction(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		t.createTransaction(w, r)

	default:
		http.Error(w, "page not found", http.StatusBadRequest)
	}
}

func (t *Transporter) createTransaction(w http.ResponseWriter, r *http.Request) {
	d := json.NewDecoder(r.Body)
	var tr TransactionRequest
	err := d.Decode(&tr)
	if err != nil {
		io.WriteString(w, string(http2.JsonStatus("fail")))
		http.Error(w, "failed to read request body", http.StatusBadRequest)
		return
	}

	if !tr.Validate() {
		io.WriteString(w, string(http2.JsonStatus("fail")))
		http.Error(w, "bad request body", http.StatusBadRequest)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	_, err = t.server.CreateTransaction(*tr.SenderPrivateKey, *tr.SenderPublicKey, *tr.SenderBlockchainAddress, *tr.RecipientBlockchainAddress, *tr.Value)
	if err != nil {
		io.WriteString(w, string(http2.JsonStatus("fail")))
		http.Error(w, "failed to read request body", http.StatusBadRequest)
		return
	}
	io.WriteString(w, string(http2.JsonStatus("success")))
}
