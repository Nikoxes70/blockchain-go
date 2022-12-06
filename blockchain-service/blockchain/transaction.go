package blockchain

import (
	"encoding/json"
	"fmt"
	"strings"
)

type Transaction struct {
	sender    string
	recipient string
	value     float32
}

func NewTransaction(sender, recipient string, value float32) *Transaction {
	return &Transaction{
		sender:    sender,
		recipient: recipient,
		value:     value,
	}
}

func (t *Transaction) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Sender    string  `json:"sender_blockchain_address"`
		Recipient string  `json:"recipient_blockchain_address"`
		Value     float32 `json:"value"`
	}{
		Sender:    t.sender,
		Recipient: t.recipient,
		Value:     t.value,
	})
}

func (t *Transaction) UnmarshalJSON(b []byte) error {
	s := struct {
		Sender    *string  `json:"sender_blockchain_address"`
		Recipient *string  `json:"recipient_blockchain_address"`
		Value     *float32 `json:"value"`
	}{
		Sender:    &t.sender,
		Recipient: &t.recipient,
		Value:     &t.value,
	}
	return json.Unmarshal(b, s)
}

func (t *Transaction) Print() {
	fmt.Printf("%s\n", strings.Repeat("-", 40))
	fmt.Printf(" sender_blockchain_address      %s\n", t.sender)
	fmt.Printf(" recipient_blockchain_address   %s\n", t.recipient)
	fmt.Printf(" value                          %.1f\n", t.value)
}

type TransactionRequest struct {
	SenderBlockchainAddress    *string  `json:"sender_blockchain_address"`
	RecipientBlockchainAddress *string  `json:"recipient_blockchain_address"`
	SenderPublicKey            *string  `json:"sender_public_key"`
	Value                      *float32 `json:"value"`
	Signature                  *string  `json:"signature"`
}

func (t *TransactionRequest) Print() {
	fmt.Printf("%s\n", strings.Repeat("-", 40))
	fmt.Printf(" sender_blockchain_address      %s\n", *t.SenderBlockchainAddress)
	fmt.Printf(" recipient_blockchain_address   %s\n", *t.RecipientBlockchainAddress)
	fmt.Printf(" sender_public_key              %s\n", *t.SenderPublicKey)
	fmt.Printf(" value                          %.1f\n", *t.Value)
	fmt.Printf(" signature                      %s\n", *t.Signature)
}

func (tr *TransactionRequest) Validate() bool {
	return tr.RecipientBlockchainAddress != nil &&
		tr.SenderBlockchainAddress != nil &&
		tr.Value != nil && tr.Signature != nil &&
		tr.SenderPublicKey != nil
}

type TransactionResponse struct {
	ID string `json:"id"`
}
