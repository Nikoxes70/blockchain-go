package blockchain_server

import (
	"blockchain/blockchain-service/blockchain"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
)

type Serverer interface {
	AddTransaction(sender, receiver string, amount float32) (id string, err error)
	GetBlockchain(address string) (*blockchain.Blockchain, error)
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
		w.Header().Add("Content-Type", "application/json")
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

func (t Transporter) HandleTransactions(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/transactions" || r.Method != http.MethodPost {
		http.Error(w, "404 not found.", http.StatusMethodNotAllowed)
		return
	}

	var reqBody map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "blockchain-server error - failed to parse request body", http.StatusInternalServerError)
		return
	}

	sender := r.URL.Query().Get("sender")
	recipient := r.URL.Query().Get("recipient")
	am := r.URL.Query().Get("amount")
	amount, err := strconv.ParseFloat(am, 32)
	if err != nil {
		http.Error(w, "blockchain-server error - failed to parse amount", http.StatusBadRequest)
		return
	}

	id, err := t.server.AddTransaction(sender, recipient, float32(amount))
	if err != nil {
		http.Error(w, "blockchain-server error - failed to create short url", http.StatusInternalServerError)
		return
	}

	responseBody := blockchain.TransactionResponse{
		ID: id,
	}

	b, err := json.Marshal(responseBody)
	if err != nil {
		http.Error(w, "blockchain-server error - something went wrong", 500)
		return
	}

	if _, err := w.Write(b); err != nil {
		fmt.Printf("failed to martshal response - %v\n", err)
	}

	return
}
