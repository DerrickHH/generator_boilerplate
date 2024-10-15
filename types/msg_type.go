package types

type AccountsMsg struct {
	Content       [][]byte `json:"content"`
	AddressNumber int      `json:"number"`
	ShardID       int      `json:"shard_id"`
}

type RequestMsg struct {
	Timestamp    int64    `json:"timestamp"`
	Transactions [][]byte `json:"transactions"`
	SequenceID   int64    `json:"sequenceID"`
}

type GenerateAccountRequest struct {
	Number  int `json:"number"`
	ShardID int `json:"shard_id"`
}

type GenerateTransactionRequest struct {
	Number          int `json:"number"`
	ShardID         int `json:"shard_id"`
	CrossShardRatio int `json:"crossShardRatio"` // 用整数代替小数
}
