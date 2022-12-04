package blockchain

import (
	"blockchain/foundation/cryptography"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"
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
	port              uint16
}

func NewBlockchain(blockchainAddress string, port uint16) (*Blockchain, error) {
	b := &Block{}
	bc := &Blockchain{
		blockchainAddress: blockchainAddress,
		port:              port,
	}

	hash, err := b.Hash()
	if err != nil {
		return nil, err
	}
	bc.createBlock(0, hash)
	return bc, nil
}

// Public

func (bc *Blockchain) TransactonPool() []*Transaction {
	return bc.transactionPool
}

func (bc *Blockchain) CreateTransaction(sender, recipient string, value float32, pKey *ecdsa.PublicKey, s *cryptography.Signature) error {
	err := bc.AddTransaction(sender, recipient, value, pKey, s)
	//Sync
	return err
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

func (bc *Blockchain) Mine() (int64, error) {
	err := bc.AddTransaction(BENEFACTOR_ADDRESS, bc.blockchainAddress, MINING_REWARD, nil, nil)
	if err != nil {
		return 0, err
	}

	nonce, err := bc.proofOfWork()
	if err != nil {
		return 0, err
	}

	prevHash, err := bc.lastBlock().Hash()
	if err != nil {
		return 0, err
	}
	b := bc.createBlock(nonce, prevHash)
	return b.Timestamp, nil
}

func (bc *Blockchain) CalculateBalance(address string) float32 {
	var balance float32 = 0
	for _, b := range bc.chain {
		for _, t := range b.Transactions {
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
		Blocks []*Block `json:"chains"`
	}{
		Blocks: bc.chain,
	})
}

// Private

func (bc *Blockchain) createBlock(nonce int, previousHash [32]byte) *Block {
	b := NewBlock(nonce, previousHash, bc.transactionPool)
	bc.chain = append(bc.chain, b)
	bc.transactionPool = []*Transaction{}
	return b
}

func (bc *Blockchain) lastBlock() *Block {
	return bc.chain[len(bc.chain)-1]
}

func (bc *Blockchain) copyTransactionPool() []*Transaction {
	transactions := make([]*Transaction, len(bc.transactionPool))
	for i, t := range bc.transactionPool {
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
		Nonce:        nonce,
		PreviousHash: prevHash,
		Timestamp:    0,
		Transactions: trs,
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
