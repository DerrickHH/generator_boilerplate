package types

import "github.com/ethereum/go-ethereum/rlp"

type Receipt struct {
	Status bool
}

func (r *Receipt) RLPEncode() ([]byte, error) {
	encoded, err := rlp.EncodeToBytes(r)
	if err != nil {
		return nil, err
	}
	return encoded, nil
}

func (r *Receipt) RLPDecode(content []byte) error {
	err := rlp.DecodeBytes(content, r)
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
