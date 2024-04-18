package server

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"generator_boilerplate/constant"
	"generator_boilerplate/generator"
	"generator_boilerplate/types"
	"log"
	"math"
	"math/big"
	"net/http"
	"strconv"
	"time"
)

var SequenceID = 1

type Server struct {
	Port string
	// NOTE: 用于生成transaction
	AddressMap map[int][]types.Account
	// NOTE: 用于记录每个 shard 的 #0 节点
	ShardsTable map[string]string
}

func NewServer(port string) *Server {
	server := &Server{
		Port:        port,
		AddressMap:  make(map[int][]types.Account),
		ShardsTable: make(map[string]string),
	}
	server.ShardsTable = constant.ShardsTable
	return server
}

func (s *Server) setRoutes() {
	http.HandleFunc("/generate_account", s.handleGenerateAccounts)
	http.HandleFunc("/generate_transaction", s.handleGenerateTransactions)
}

func (s *Server) handleGenerateAccounts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	params := r.URL.Query()
	// NOTE: 知道是哪个 shard 需要生成 多少个 accounts
	param1 := params.Get("shard_id")
	param2 := params.Get("acc_number")

	shardID, _ := strconv.Atoi(param1)
	accNumber, _ := strconv.Atoi(param2)

	accounts, err := generator.GenerateAccounts(accNumber)
	if err != nil {
		http.Error(w, "Error generating accounts", http.StatusInternalServerError)
		return
	}
	s.AddressMap[shardID] = accounts
	log.Println("Generated Accounts.")

	msg := types.AccountsMsg{}
	msg.Content = make([][]byte, len(accounts))
	for i := 0; i < len(accounts); i++ {
		msg.Content[i], _ = accounts[i].RLPEncode()
	}
	msg.AddressNumber = accNumber
	jsonData, err := json.Marshal(msg)
	if err != nil {
		log.Fatalf("Failed to JSON marshal account: %v", err)
	}

	resp, err := http.Post(s.ShardsTable[fmt.Sprintf("Shard_%d", shardID)]+"/accounts", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatalf("Failed to send account to shard: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Failed to send account to shard %d, received status code: %d", shardID, resp.StatusCode)
	} else {
		fmt.Printf("%d accounts sent to shard %d successfully\n", accNumber, shardID)
	}
}

func (s *Server) handleGenerateTransactions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}
	params := r.URL.Query()
	param1 := params.Get("shard_id")
	param2 := params.Get("is_overload")
	shardID, _ := strconv.Atoi(param1)
	isOverload, _ := strconv.ParseBool(param2)

	ticker := time.NewTicker(10 * time.Second)
	go func() {
		for range ticker.C {
			// NOTE: 默认生成 1000 笔交易
			number := constant.TransactionsGeneration
			if isOverload == true {
				number = int(math.Round(float64(number) * (1 + constant.OverloadTransactionsRatio)))
			}
			log.Println("========== Generating Transactions ==========")
			generatedTransactions := make([]interface{}, 0)
			// NOTE: 增加一个计数器，保证交易的分散性
			counter := make(map[string]int)
			// NOTE: 控制交易重复
			repetitive := make(map[string][]string)
			// NOTE: 控制 nonce
			noncer := make(map[string]int64)
			for _, acc := range s.AddressMap[shardID] {
				counter[acc.Address] = 0
				repetitive[acc.Address] = make([]string, 0)
				// NOTE: nonce 的同步也是一个问题
				noncer[acc.Address] = acc.Nonce + 1
			}
			trans, ctrans := 0, 0
			for trans+ctrans < number {
				rnd, _ := rand.Int(rand.Reader, big.NewInt(100))
				if int(rnd.Int64()) > constant.CrossShardTransactionRatio {
					tx, err := generator.GenerateTransaction(s.AddressMap[shardID], &counter, &repetitive, &noncer)
					if err != nil {
						log.Println("[ERROR] Wrong when generating the transactions: ", err)
						continue
					}
					generatedTransactions = append(generatedTransactions, tx)
					trans += 1
				} else {
					ctx, err := generator.GenerateCrossShardTransaction(shardID, s.AddressMap, &counter, &repetitive, &noncer)
					if err != nil {
						log.Println("[ERROR] Wrong when generating the cross shard transactions: ", err)
						continue
					}
					generatedTransactions = append(generatedTransactions, ctx)
					ctrans += 1
				}
			}
			log.Println("========== Generated Transactions ==========")

			msg := types.RequestMsg{}
			msg.Timestamp = time.Now().UnixNano()
			msg.Transactions = make([][]byte, trans)
			msg.CrossShardTransactions = make([][]byte, ctrans)
			for i := 0; i < len(generatedTransactions); i++ {
				switch generatedTransactions[i].(type) {
				case types.Transaction:
					msg.Transactions[i], _ = generatedTransactions[i].(*types.Transaction).Marshal()
				case types.CrossShardTransaction:
					msg.CrossShardTransactions[i], _ = generatedTransactions[i].(*types.CrossShardTransaction).RLPEncode()
				}
			}
			msg.SequenceID = int64(SequenceID)
			SequenceID++
			msg.TransactionNumber = len(generatedTransactions)
			jsonData, err := json.Marshal(msg)
			if err != nil {
				log.Fatalf("Failed to JSON marshal account: %v", err)
			}

			resp, err := http.Post(s.ShardsTable[fmt.Sprintf("Shard_%d", shardID)]+"/req", "application/json", bytes.NewBuffer(jsonData))
			if err != nil {
				log.Fatalf("Failed to send transactions to shard: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				log.Printf("Failed to send transactions to shard %d, received status code: %d", shardID, resp.StatusCode)
			} else {
				fmt.Printf("%d transactions sent to shard %d successfully\n", len(generatedTransactions), shardID)
			}
		}
	}()
}

func (s *Server) Start() {
	s.setRoutes()
	fmt.Printf("Server is running on http://0.0.0.0:%s/\n", s.Port)
	err := http.ListenAndServe("0.0.0.0:"+s.Port, nil)
	if err != nil {
		log.Fatal(err)
	}
}
