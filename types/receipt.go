package types

import (
	"encoding/json"
)

type Receipt struct {
	Status bool `json:"status"`
}

//func (r *Receipt) RLPEncode() ([]byte, error) {
//	encoded, err := rlp.EncodeToBytes(r)
//	if err != nil {
//		return nil, err
//	}
//	return encoded, nil
//}
//
//func (r *Receipt) RLPDecode(content []byte) error {
//	err := rlp.DecodeBytes(content, r)
//	if err != nil {
//		return err
//	}
//	return nil
//}

func (r *Receipt) Marshal() ([]byte, error) {
	encoded, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}
	return encoded, nil
}

func (r *Receipt) Unmarshal(content []byte) error {
	err := json.Unmarshal(content, r)
	if err != nil {
		return err
	}
	return nil
}

func (r *Receipt) SetStatus(s bool) {
	r.Status = s
}

func (r *Receipt) GetStatus() bool {
	return r.Status
}
