package types

import (
	"encoding/json"
)

type Account struct {
	PrivateKey string `json:"private_key"`
	Address    string `json:"address"`
	Balance    int64  `json:"balance"`
	Nonce      int64  `json:"nonce"`
}

//func (a *Account) RLPEncode() ([]byte, error) {
//	encoded, err := rlp.EncodeToBytes(a)
//	fmt.Println(string(encoded))
//	if err != nil {
//		return nil, err
//	}
//	return encoded, nil
//}

func (a *Account) Marshal() ([]byte, error) {
	encoded, err := json.Marshal(a)
	if err != nil {
		return nil, err
	}
	return encoded, nil
}

//func (a *Account) RLPDecode(content []byte) error {
//	err := rlp.DecodeBytes(content, a)
//	if err != nil {
//		return err
//	}
//	return nil
//}

func (a *Account) Unmarshal(content []byte) error {
	err := json.Unmarshal(content, a)
	if err != nil {
		return err
	}
	return nil
}
