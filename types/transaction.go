package types

import (
	"crypto/sha256"
	"encoding/json"
	"generator_boilerplate/utils"
	"math/big"
	"time"
)

type Transaction struct {
	From      utils.Address `json:"from"`
	To        utils.Address `json:"to"`
	Value     *big.Int      `json:"value"`
	Signature []byte        `json:"signature"`
	Nonce     uint64        `json:"nonce"`
	Hash      []byte        `json:"hash"`
	Time      time.Time     `json:"time"`
	// used in transaction relaying
	Relayed bool `json:"relayed"`
}

func NewTransaction(from, to string, value *big.Int, nonce uint64) *Transaction {
	tx := &Transaction{
		From:  from,
		To:    to,
		Value: value,
		Nonce: nonce,
		Time:  time.Now(),
	}

	txMar, _ := json.Marshal(tx)
	hash := sha256.Sum256(txMar)
	tx.Hash = hash[:]
	tx.Relayed = false
	return tx
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
