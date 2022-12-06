package blockchain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
)

type Block struct {
	timestamp    int64
	nonce        int
	previousHash [32]byte
	transactions []*Transaction
}

func NewBlock(nonce int, previousHash [32]byte, ts []*Transaction) *Block {
	return &Block{
		timestamp:    time.Now().UnixNano(),
		nonce:        nonce,
		previousHash: previousHash,
		transactions: ts,
	}
}

func (b *Block) GetTimestamp() int64 {
	return b.timestamp
}

func (b *Block) GetPreviousHash() [32]byte {
	return b.previousHash
}

func (b *Block) GetNonce() int {
	return b.nonce
}

func (b *Block) GetTransactions() []*Transaction {
	return b.transactions
}

func (b *Block) Hash() ([32]byte, error) {
	bts, err := json.Marshal(b)
	if err != nil {
		return [32]byte{}, err
	}
	return sha256.Sum256(bts), nil
}

func (b *Block) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Nonce        int            `json:"nonce"`
		PreviousHash string         `json:"previous_hash"`
		Timestamp    int64          `json:"timestamp"`
		Transactions []*Transaction `json:"transactions"`
	}{
		Nonce:        b.nonce,
		PreviousHash: fmt.Sprintf("%x", b.previousHash),
		Timestamp:    b.timestamp,
		Transactions: b.transactions,
	})
}

func (b *Block) UnmarshalJSON(bts []byte) error {
	var previousHash string
	s := struct {
		Nonce        *int            `json:"nonce"`
		PreviousHash *string         `json:"previous_hash"`
		Timestamp    *int64          `json:"timestamp"`
		Transactions *[]*Transaction `json:"transactions"`
	}{
		Nonce:        &b.nonce,
		PreviousHash: &previousHash,
		Timestamp:    &b.timestamp,
		Transactions: &b.transactions,
	}
	if err := json.Unmarshal(bts, &s); err != nil {
		return err
	}

	ph, err := hex.DecodeString(*s.PreviousHash)
	if err != nil {
		return err
	}
	copy(b.previousHash[:], ph[:32])
	return nil
}

func (b *Block) Print() {
	fmt.Printf("timestamp       %d\n", b.timestamp)
	fmt.Printf("nonce           %d\n", b.nonce)
	fmt.Printf("previous_hash   %x\n", b.previousHash)
	//fmt.Printf("transactions    %v\n", b.Transactions)
	for _, t := range b.transactions {
		t.Print()
	}
}
