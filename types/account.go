package types

import (
	"encoding/json"
)

type Account struct {
	PrivateKey string `json:"private_key"`
	Address    string `json:"address"`
	Balance    int64  `json:"balance"`
	IntraNonce int64  `json:"intra_nonce"`
	InterNonce int64  `json:"inter_nonce"`
}

func (a *Account) Marshal() ([]byte, error) {
	encoded, err := json.Marshal(a)
	if err != nil {
		return nil, err
	}
	return encoded, nil
}

func (a *Account) Unmarshal(content []byte) error {
	err := json.Unmarshal(content, a)
	if err != nil {
		return err
	}
	return nil
}
