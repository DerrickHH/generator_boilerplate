package types

import (
	"encoding/json"
	"generator_boilerplate/utils"
	"math/big"
)

type AccountState struct {
	Address   utils.Address
	PublicKey []byte
	Nonce     uint64
	Balance   *big.Int
	// Only used for smart contracts
	StorageRoot []byte
	CodeHash    []byte
}

func (a *AccountState) Marshal() ([]byte, error) {
	encoded, err := json.Marshal(a)
	if err != nil {
		return nil, err
	}
	return encoded, nil
}

func (a *AccountState) Unmarshal(content []byte) error {
	err := json.Unmarshal(content, a)
	if err != nil {
		return err
	}
	return nil
}
