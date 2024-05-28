package types

import (
	"crypto/sha256"
	"encoding/json"
)

type Transaction struct {
	From    string  `json:"from"`
	To      string  `json:"to"`
	Value   int64   `json:"value"`
	Nonce   int64   `json:"nonce"`
	Receipt Receipt `json:"receipt"`
	Hash    []byte  `json:"hash"`
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

func (t *Transaction) Marshal() ([]byte, error) {
	encoded, err := json.Marshal(t)
	if err != nil {
		return nil, err
	}
	return encoded, nil
}

func (t *Transaction) Unmarshal(content []byte) error {
	err := json.Unmarshal(content, t)
	if err != nil {
		return err
	}
	return nil
}
