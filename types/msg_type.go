package types

type AccountsMsg struct {
	Content       [][]byte `json:"content"`
	AddressNumber int      `json:"number"`
}

type RequestMsg struct {
	Timestamp         int64 `json:"timestamp"`
	TransactionNumber int   `json:"number"`
	// ClientID   string `json:"clientID"`
	Transactions           [][]byte `json:"transactions"`
	CrossShardTransactions [][]byte `json:"cross_shard_transaction"`
	SequenceID             int64    `json:"sequenceID"`
}
