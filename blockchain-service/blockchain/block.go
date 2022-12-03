package blockchain

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"
)

type Block struct {
	Nonce        int
	PreviousHash [32]byte
	Timestamp    int64
	Transactions []*Transaction
}

func NewBlock(nonce int, previousHash [32]byte, ts []*Transaction) *Block {
	return &Block{
		Timestamp:    time.Now().UnixNano(),
		Nonce:        nonce,
		PreviousHash: previousHash,
		Transactions: ts,
	}
}

func (b *Block) MarshallJSON() ([]byte, error) {
	return json.Marshal(struct {
		Nonce        int            `json:"nonce"`
		PreviousHash string         `json:"previous_hash"`
		Timestamp    int64          `json:"timestamp"`
		Transactions []*Transaction `json:"transactions"`
	}{
		Nonce:        b.Nonce,
		PreviousHash: fmt.Sprintf("%x", b.PreviousHash),
		Timestamp:    b.Timestamp,
		Transactions: b.Transactions,
	})
}

func (b *Block) Hash() ([32]byte, error) {
	bts, err := json.Marshal(b)
	if err != nil {
		return [32]byte{}, err
	}
	return sha256.Sum256(bts), nil
}

func (b *Block) Print() {
	fmt.Printf("timestamp       %d\n", b.Timestamp)
	fmt.Printf("nonce           %d\n", b.Nonce)
	fmt.Printf("previous_hash   %x\n", b.PreviousHash)
	//fmt.Printf("transactions    %v\n", b.Transactions)
	for _, t := range b.Transactions {
		t.Print()
	}
}
