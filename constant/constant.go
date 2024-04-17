package constant

const (
	Port                       = "8000"
	Balance                    = 10000000
	TransactionsGeneration     = 1000
	OverloadTransactionsRatio  = 0.25
	CrossShardTransactionRatio = 25
	MaxTxsInBlock              = 20
)

var ShardsTable = map[string]string{
	"Shard_1": "http://127.0.0.1:6000",
	"Shard_2": "http://127.0.0.1:7000",
	"Shard_3": "http://127.0.0.1:8000",
}
