package types

import "github.com/ethereum/go-ethereum/rlp"

type Account struct {
	PrivateKey string `json:"private_key"`
	Address    string `json:"address"`
	Balance    int64  `json:"balance"`
	Nonce      int64  `json:"nonce"`
	// TODO: 暂时不明确
	States int `json:"states"`
	// TODO: 用于后续分身，表明该账户都出现过哪些 shard 中
	ShardList []int `json:"shard_list"`
}

func (a *Account) RLPEncode() ([]byte, error) {
	encoded, err := rlp.EncodeToBytes(a)
	if err != nil {
		return nil, err
	}
	return encoded, nil
}

func (a *Account) RLPDecode(content []byte) error {
	err := rlp.DecodeBytes(content, a)
	if err != nil {
		return err
	}
	return nil
}
