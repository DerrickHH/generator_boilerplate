package types

import (
	"crypto/sha256"
	"encoding/json"
	"github.com/ethereum/go-ethereum/rlp"
)

type Transaction struct {
	From    string
	To      string
	Value   int64
	Nonce   int64
	Receipt Receipt
	Hash    []byte
}

func NewTransaction(from, to string, value, nonce int64) Transaction {
	return Transaction{
		From:    from,
		To:      to,
		Value:   value,
		Nonce:   nonce,
		Receipt: Receipt{},
	}
}

func (t *Transaction) GenerateTransactionHash() error {
	txData := Transaction{
		From:  t.From,
		To:    t.To,
		Value: t.Value,
		Nonce: t.Nonce,
	}

	data, err := json.Marshal(txData)
	if err != nil {
		return err
	}

	hash := sha256.Sum256(data)
	t.Hash = hash[:]
	return nil
}

func (t *Transaction) RLPEncode() ([]byte, error) {
	encoded, err := rlp.EncodeToBytes(t)
	if err != nil {
		return nil, err
	}
	return encoded, nil
}

func (t *Transaction) RLPDecode(content []byte) error {
	err := rlp.DecodeBytes(content, t)
	if err != nil {
		return err
	}
	return nil
}
