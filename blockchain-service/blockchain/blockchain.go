package blockchain

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"blockchain/foundation/cryptography"
)

const (
	MIN_DIFFICULTY     = 2
	BENEFACTOR_ADDRESS = "THE BLOCKCHAIN"
	MINING_REWARD      = 0.0001
)

type Blockchain struct {
	transactionPool   []*Transaction
	chain             []*Block
	blockchainAddress string
	mux               sync.Mutex
}

func NewBlockchain(blockchainAddress string) (*Blockchain, error) {
	b := &Block{}
	bc := &Blockchain{
		blockchainAddress: blockchainAddress,
	}

	hash, err := b.Hash()
	if err != nil {
		return nil, err
	}
	bc.createBlock(0, hash)
	return bc, nil
}

// Public

func (bc *Blockchain) SetChain(c []*Block) {
	bc.chain = c
}

func (bc *Blockchain) Chain() []*Block {
	return bc.chain
}

func (bc *Blockchain) TransactionPool() []*Transaction {
	return bc.transactionPool
}

func (bc *Blockchain) TruncateTransactionPool() int {
	l := len(bc.transactionPool)
	bc.transactionPool = []*Transaction{}
	return l
}

func (bc *Blockchain) CreateTransaction(sender, recipient string, value float32, pKey *ecdsa.PublicKey, s *cryptography.Signature) error {
	return bc.AddTransaction(sender, recipient, value, pKey, s)
}

func (bc *Blockchain) AddTransaction(sender, recipient string, value float32, pKey *ecdsa.PublicKey, s *cryptography.Signature) error {
	t := NewTransaction(sender, recipient, value)

	if sender != BENEFACTOR_ADDRESS {
		valid, err := bc.verifyTransactionSignature(pKey, s, t)
		if err != nil {
			return err
		}
		if !valid {
			return fmt.Errorf("invalid transaction signature")
		}
		//if bc.CalculateBalance(sender) < value {
		//	return fmt.Errorf("not enouth funds")
		//}
	}

	bc.transactionPool = append(bc.transactionPool, t)
	return nil
}

func (bc *Blockchain) Mine() (int64, bool, error) {
	bc.mux.Lock()
	defer bc.mux.Unlock()

	if len(bc.TransactionPool()) == 0 {
		return 0, false, nil
	}

	err := bc.AddTransaction(BENEFACTOR_ADDRESS, bc.blockchainAddress, MINING_REWARD, nil, nil)
	if err != nil {
		return 0, false, err
	}

	nonce, err := bc.proofOfWork()
	if err != nil {
		return 0, false, err
	}

	prevHash, err := bc.lastBlock().Hash()
	if err != nil {
		return 0, false, err
	}
	b := bc.createBlock(nonce, prevHash)
	return b.timestamp, true, nil
}

func (bc *Blockchain) CalculateBalance(address string) float32 {
	var balance float32 = 0
	for _, b := range bc.chain {
		for _, t := range b.GetTransactions() {
			if t.recipient == address {
				balance += t.value
			}

			if t.sender == address {
				balance -= t.value
			}
		}
	}
	return balance
}

func (bc *Blockchain) ValidChain(chain []*Block) (bool, error) {
	preBlock := chain[0]
	currentIndex := 1

	for currentIndex < len(chain) {
		b := chain[currentIndex]
		hash, err := preBlock.Hash()
		if err != nil {
			return false, err
		}

		if b.GetPreviousHash() != hash {
			return false, nil
		}

		valid, err := bc.validProof(b.GetNonce(), b.GetPreviousHash(), b.GetTransactions(), MIN_DIFFICULTY)
		if err != nil {
			return false, err
		}

		if !valid {
			return false, nil
		}

		preBlock = b
		currentIndex++
	}

	return true, nil
}

func (bc *Blockchain) Print() {
	for i, block := range bc.chain {
		fmt.Printf("%s chain %d %s\n", strings.Repeat("=", 25), i,
			strings.Repeat("=", 25))
		block.Print()
	}
	fmt.Printf("%s\n", strings.Repeat("*", 25))
}

func (bc *Blockchain) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Blocks []*Block `json:"chain"`
	}{
		Blocks: bc.chain,
	})
}

func (bc *Blockchain) UnmarshalJSON(b []byte) error {
	s := struct {
		Blocks *[]*Block `json:"chain"`
	}{
		Blocks: &bc.chain,
	}
	return json.Unmarshal(b, &s)
}

// Private

func (bc *Blockchain) createBlock(nonce int, previousHash [32]byte) *Block {
	b := NewBlock(nonce, previousHash, bc.TransactionPool())
	bc.chain = append(bc.chain, b)
	bc.transactionPool = []*Transaction{}
	return b
}

func (bc *Blockchain) lastBlock() *Block {
	return bc.chain[len(bc.chain)-1]
}

func (bc *Blockchain) copyTransactionPool() []*Transaction {
	transactions := make([]*Transaction, len(bc.TransactionPool()))
	for i, t := range bc.TransactionPool() {
		transactions[i] = NewTransaction(t.sender, t.recipient, t.value)
	}
	return transactions
}

func (bc *Blockchain) verifyTransactionSignature(sender *ecdsa.PublicKey, sign *cryptography.Signature, t *Transaction) (bool, error) {
	b, err := json.Marshal(t)
	if err != nil {
		return false, err
	}
	h := sha256.Sum256(b)
	return ecdsa.Verify(sender, h[:], sign.R, sign.S), nil
}

func (bc *Blockchain) validProof(nonce int, prevHash [32]byte, trs []*Transaction, difficulty int) (bool, error) {
	zeros := strings.Repeat("0", difficulty)
	guessBlock := Block{
		nonce:        nonce,
		previousHash: prevHash,
		timestamp:    0,
		transactions: trs,
	}
	hash, err := guessBlock.Hash()
	if err != nil {
		return false, nil
	}
	gHashString := fmt.Sprintf("%x", hash)
	return gHashString[:difficulty] == zeros, nil
}

func (bc *Blockchain) proofOfWork() (int, error) {
	trs := bc.copyTransactionPool()
	prevHash, err := bc.lastBlock().Hash()
	if err != nil {
		return 0, err
	}

	nonce := 0
	invalid := true
	for invalid {
		valid, err := bc.validProof(nonce, prevHash, trs, MIN_DIFFICULTY)
		if err != nil {
			return 0, err
		}
		invalid = !valid
		nonce += 1
	}
	return nonce, nil
}
