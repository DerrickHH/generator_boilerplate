package types

import (
	"crypto/sha256"
	"encoding/json"
	"github.com/ethereum/go-ethereum/rlp"
)

type CrossShardTransaction struct {
	ShardID int
	From    string
	To      string
	Value   int64
	Nonce   int64
	Receipt Receipt
	Hash    []byte
	// NOTE: proof 字段在填充前需要先 json 编码
	Proof []byte
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

func (cst *CrossShardTransaction) RLPEncode() ([]byte, error) {
	encoded, err := rlp.EncodeToBytes(cst)
	if err != nil {
		return nil, err
	}
	return encoded, nil
}

func (cst *CrossShardTransaction) RLPDecode(content []byte) error {
	err := rlp.DecodeBytes(content, cst)
	if err != nil {
		return nil
	}
	return nil
}
