package types

import (
	"crypto/sha256"
	"encoding/json"
)

type CrossShardTransaction struct {
	ShardID int     `json:"shard_id"`
	From    string  `json:"from"`
	To      string  `json:"to"`
	Value   int64   `json:"value"`
	Nonce   int64   `json:"nonce"`
	Receipt Receipt `json:"receipt"`
	Hash    []byte  `json:"hash"`
	// NOTE: proof 字段在填充前需要先 json 编码
	Proof []byte `json:"proof"`
}

func NewCrossShardTransaction(shardID int, from, to string, value, nonce int64) CrossShardTransaction {
	return CrossShardTransaction{
		ShardID: shardID,
		From:    from,
		To:      to,
		Value:   value,
		Nonce:   nonce,
		Receipt: Receipt{},
	}
}

func (cst *CrossShardTransaction) GenerateTransactionHash() error {
	txData := CrossShardTransaction{
		ShardID: cst.ShardID,
		From:    cst.From,
		To:      cst.To,
		Value:   cst.Value,
		Nonce:   cst.Nonce,
	}

	data, err := json.Marshal(txData)
	if err != nil {
		return err
	}

	hash := sha256.Sum256(data)
	cst.Hash = hash[:]
	return nil
}

//func (cst *CrossShardTransaction) RLPEncode() ([]byte, error) {
//	encoded, err := rlp.EncodeToBytes(cst)
//	if err != nil {
//		return nil, err
//	}
//	return encoded, nil
//}
//
//func (cst *CrossShardTransaction) RLPDecode(content []byte) error {
//	err := rlp.DecodeBytes(content, cst)
//	if err != nil {
//		return nil
//	}
//	return nil
//}

func (cst *CrossShardTransaction) Marshal() ([]byte, error) {
	encoded, err := json.Marshal(cst)
	if err != nil {
		return nil, err
	}
	return encoded, nil
}

func (cst *CrossShardTransaction) Unmarshal(content []byte) error {
	err := json.Unmarshal(content, cst)
	if err != nil {
		return err
	}
	return nil
}
