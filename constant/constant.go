package constant

const (
	Port                       = "8000"
	Balance                    = 10000000
	TransactionsGeneration     = 10
	OverloadTransactionsRatio  = 0.25
	CrossShardTransactionRatio = 25
	MaxTxsInBlock              = 20
)

var ShardsTable = map[string]string{
	"Shard_0": "http://127.0.0.1:9200",
	"Shard_1": "http://127.0.0.1:10200",
	"Shard_2": "http://127.0.0.1:8000",
}
