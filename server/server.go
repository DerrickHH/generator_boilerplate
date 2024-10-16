package server

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"generator_boilerplate/constant"
	"generator_boilerplate/generator"
	"generator_boilerplate/types"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
	"log"
	"math/big"
	"net/http"
	"reflect"
	"time"
)

var SequenceID = 1

type Server struct {
	Url string
	// NOTE: 用于生成transaction
	AddressMap map[int]map[uint64]*types.AccountState
	// NOTE: 用于记录每个 shard 的 #0 节点
	ShardsTable     map[string]string
	BeaconNode      string
	AccountsKeyMap  map[int]map[uint64]types.SignatureAccount
	AccountsMsg     []*types.AccountsMsg
	TransactionsMsg []*types.RequestMsg

	MsgBuffer   *MsgBuffer
	MsgEntrance chan interface{}
	MsgDelivery chan interface{}
}

type MsgBuffer struct {
	GenerateAccountsRequests     []*types.GenerateAccountRequest
	GenerateTransactionsRequests []*types.GenerateTransactionRequest
}

func NewServer(url string) *Server {
	server := &Server{
		Url:             url + ":" + constant.Port,
		AddressMap:      make(map[int]map[uint64]*types.AccountState),
		ShardsTable:     make(map[string]string),
		BeaconNode:      constant.BeaconNode,
		AccountsMsg:     make([]*types.AccountsMsg, 0),
		TransactionsMsg: make([]*types.RequestMsg, 0),
		AccountsKeyMap:  make(map[int]map[uint64]types.SignatureAccount),
		MsgBuffer: &MsgBuffer{
			GenerateAccountsRequests:     make([]*types.GenerateAccountRequest, 0),
			GenerateTransactionsRequests: make([]*types.GenerateTransactionRequest, 0),
		},
		MsgEntrance: make(chan interface{}),
		MsgDelivery: make(chan interface{}),
	}
	server.ShardsTable = constant.ShardsTable
	go server.dispatchMsg()
	go server.resolveMsg()
	server.setRoutes()
	return server
}

func (s *Server) setRoutes() {
	http.HandleFunc("/generate_account", s.getGenerateAccountRequest)
	http.HandleFunc("/generate_transaction", s.getGenerateTransactionRequest)
}

func (s *Server) getGenerateAccountRequest(w http.ResponseWriter, r *http.Request) {
	var msg *types.GenerateAccountRequest
	err := json.NewDecoder(r.Body).Decode(&msg)
	if err != nil {
		log.Println(err)
		return
	}
	s.MsgEntrance <- &msg
}

func (s *Server) getGenerateTransactionRequest(w http.ResponseWriter, r *http.Request) {
	var msg *types.GenerateTransactionRequest
	err := json.NewDecoder(r.Body).Decode(&msg)
	if err != nil {
		log.Println(err)
		return
	}
	s.MsgEntrance <- &msg
}

func (s *Server) dispatchMsg() {
	for {
		msg := <-s.MsgEntrance
		err := s.routeMsg(msg)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func (s *Server) routeMsg(msg interface{}) error {
	switch msg.(type) {
	case *types.GenerateAccountRequest:
		msgs := make([]*types.GenerateAccountRequest, len(s.MsgBuffer.GenerateAccountsRequests))
		copy(msgs, s.MsgBuffer.GenerateAccountsRequests)
		msgs = append(msgs, msg.(*types.GenerateAccountRequest))
		s.MsgBuffer.GenerateAccountsRequests = make([]*types.GenerateAccountRequest, 0)
		s.MsgDelivery <- msgs

	case *types.GenerateTransactionRequest:
		msgs := make([]*types.GenerateTransactionRequest, len(s.MsgBuffer.GenerateTransactionsRequests))
		copy(msgs, s.MsgBuffer.GenerateTransactionsRequests)
		msgs = append(msgs, msg.(*types.GenerateTransactionRequest))
		s.MsgBuffer.GenerateTransactionsRequests = make([]*types.GenerateTransactionRequest, 0)
		s.MsgDelivery <- msgs
	}
	return nil
}

func (s *Server) resolveMsg() {
	for {
		msgs := <-s.MsgDelivery
		switch msgs.(type) {
		case []*types.GenerateAccountRequest:
			err := s.resolveGenerateAccountRequest(msgs.([]*types.GenerateAccountRequest))
			if err != nil {
				log.Println("Wrong when resolve the account message: ", err)
			}
		case []*types.GenerateTransactionRequest:
			err := s.resolveGenerateTransactionRequest(msgs.([]*types.GenerateTransactionRequest))
			if err != nil {
				log.Println("Wrong when resolve the transaction message: ", err)
			}
		}
	}
}

func (s *Server) resolveGenerateAccountRequest(msgs []*types.GenerateAccountRequest) []error {
	errs := make([]error, 0)
	for _, msg := range msgs {
		err := s.GenerateAccount(msg)
		if err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return errs
	}
	return nil
}

func (s *Server) resolveGenerateTransactionRequest(msgs []*types.GenerateTransactionRequest) []error {
	errs := make([]error, 0)
	for _, msg := range msgs {
		err := s.GenerateTransaction(msg)
		if err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return errs
	}
	return nil
}

func (s *Server) GenerateAccount(msg *types.GenerateAccountRequest) error {
	accounts, accountsMap := generator.GenerateAccounts(msg.Number)
	s.AccountsKeyMap[msg.ShardID] = accountsMap
	accountsMsg := &types.AccountsMsg{
		Content: make([][]byte, 0),
		ShardID: msg.ShardID,
	}
	s.AddressMap[msg.ShardID] = make(map[uint64]*types.AccountState)
	for _, acc := range accounts {
		accMar := acc.Marshal()
		accountsMsg.Content = append(accountsMsg.Content, accMar)
		s.AddressMap[msg.ShardID][acc.Index] = acc
	}
	accountsMsg.AddressNumber = len(s.AddressMap[msg.ShardID])
	jsonMsg, _ := json.Marshal(accountsMsg)
	send(s.BeaconNode+"/accounts", jsonMsg)
	send(s.ShardsTable[fmt.Sprintf("Shard_%d", msg.ShardID)]+"/accounts", jsonMsg)
	log.Println("Generated Accounts.")
	return nil
}

func (s *Server) GenerateTransaction(req *types.GenerateTransactionRequest) error {
	number := req.Number
	log.Println("========== Generating Transactions ==========")
	generatedTransactions := make([]types.Transaction, 0)
	// NOTE: 增加一个计数器，保证交易的分散性
	counter := make(map[uint64]int)
	// NOTE: 控制交易重复
	repetitive := make(map[uint64][]uint64)
	// NOTE: 控制 nonce
	noncer := make(map[uint64]int64)
	accs := s.AddressMap[req.ShardID]
	for k, _ := range accs {
		counter[k] = 0
		repetitive[k] = make([]uint64, 0)
	}
	trans, ctrans := 0, 0
	for trans+ctrans < number {
		rnd, _ := rand.Int(rand.Reader, big.NewInt(100))
		if int(rnd.Int64()) > req.CrossShardRatio {
			tx, err := s.CreateTransaction(s.AddressMap[req.ShardID], &counter, &repetitive, &noncer)
			if err != nil {
				log.Println("[ERROR] Wrong when generating the transactions: ", err)
				continue
			}
			generatedTransactions = append(generatedTransactions, *tx)
			trans += 1
		} else {
			ctx, err := s.CreateCrossShardTransaction(req.ShardID, &counter, &repetitive, &noncer)
			if err != nil {
				log.Println("[ERROR] Wrong when generating the cross shard transactions: ", err)
				continue
			}
			generatedTransactions = append(generatedTransactions, *ctx)
			ctrans += 1
		}
	}
	log.Println("========== Generated Transactions ==========")

	msg := types.RequestMsg{}
	msg.Timestamp = time.Now().UnixNano()
	msg.Transactions = make([][]byte, 0)
	for i := 0; i < len(generatedTransactions); i++ {
		transaction, _ := generatedTransactions[i].Marshal()
		msg.Transactions = append(msg.Transactions, transaction)
	}
	msg.SequenceID = int64(SequenceID)
	SequenceID++
	jsonData, err := json.Marshal(msg)
	fmt.Println(string(jsonData))
	if err != nil {
		log.Fatalf("Failed to JSON marshal account: %v", err)
	}

	send(s.ShardsTable[fmt.Sprintf("Shard_%d", req.ShardID)]+"/req", jsonData)
	return nil
}

func (s *Server) Start() {
	s.setRoutes()
	fmt.Printf("Server is running on http://0.0.0.0:%s/\n", s.Url)
	err := http.ListenAndServe("http://0.0.0.0:"+s.Url, nil)
	if err != nil {
		log.Fatal(err)
	}
}

func (s *Server) CreateTransaction(addresses map[uint64]*types.AccountState, counter *map[uint64]int, repetitive *map[uint64][]uint64, noncer *map[uint64]int64) (*types.Transaction, error) {
	indexFrom, indexTo := uint64(0), uint64(0)
	// NOTE: 严格控制 account 的 index 即可
	for indexFrom == indexTo {
		indexFromInt64, _ := rand.Int(rand.Reader, big.NewInt(int64(len(addresses))))
		indexFrom = indexFromInt64.Uint64()
		indexToInt64, _ := rand.Int(rand.Reader, big.NewInt(int64(len(addresses))))
		indexTo = indexToInt64.Uint64()
	}
	if containsUint64(addresses[indexTo].Index, (*repetitive)[addresses[indexFrom].Index]) {
		return &types.Transaction{}, errors.New("repetitive from and to")
	}
	if (*counter)[addresses[indexFrom].Index] >= constant.MaxTxsInBlock {
		return &types.Transaction{}, errors.New("transaction counter has exceed")
	}
	minBalance := fr.NewElement(1)
	if addresses[indexFrom].Balance.Cmp(&minBalance) == -1 {
		return &types.Transaction{}, errors.New("no sufficient balance")
	}
	newTx := types.NewTransaction(uint64(1), addresses[indexFrom].PublicKey, addresses[indexTo].PublicKey, 0)
	if reflect.DeepEqual(newTx.GenerateTransactionHash(mimc.NewMiMC()), []byte("")) {
		// log.Error("[ERROR] Wrong when generate the transaction: nil hash.")
		return &types.Transaction{}, errors.New("wrong tx hash")
	}
	(*repetitive)[addresses[indexFrom].Index] = append((*repetitive)[addresses[indexFrom].Index], addresses[indexTo].Index)
	(*noncer)[addresses[indexFrom].Index] += 1
	(*counter)[addresses[indexFrom].Index] += 1
	newTx.Time = time.Now()
	return &newTx, nil
}

func (s *Server) CreateCrossShardTransaction(shardID int, counter *map[uint64]int, repetitive *map[uint64][]uint64, noncer *map[uint64]int64) (*types.Transaction, error) {
	fmt.Println("The length of addressMap is: ", len(s.AddressMap))
	txIndexFromInt64, _ := rand.Int(rand.Reader, big.NewInt(int64(len(s.AddressMap[shardID]))))
	txIndexFrom := txIndexFromInt64.Uint64()
	if (*counter)[s.AddressMap[shardID][txIndexFrom].Index] >= constant.MaxTxsInBlock {
		return &types.Transaction{}, errors.New("counter has exceed")
	}
	minBalance := fr.NewElement(1)
	if s.AddressMap[shardID][txIndexFrom].Balance.Cmp(&minBalance) == -1 {
		return &types.Transaction{}, errors.New("no sufficient balance")
	}
	indexTo := shardID
	for indexTo == shardID {
		// 根据 全局 的 constant 计算 目标 shard 是谁
		indexToInt64, _ := rand.Int(rand.Reader, big.NewInt(int64(len(s.AddressMap))))
		indexTo = int(indexToInt64.Int64())
	}
	fmt.Println("indexTo is: ", indexTo)
	txIndexToInt64, _ := rand.Int(rand.Reader, big.NewInt(int64(len(s.AddressMap[indexTo]))))
	txIndexTo := txIndexToInt64.Uint64()
	// 如果这对组合的交易已经存在的，也不能保留
	if containsUint64(s.AddressMap[indexTo][txIndexTo].Index, (*repetitive)[s.AddressMap[shardID][txIndexFrom].Index]) {
		return &types.Transaction{}, errors.New("repetitive from and to")
	}
	newTx := types.NewTransaction(uint64(1), s.AddressMap[shardID][txIndexFrom].PublicKey, s.AddressMap[indexTo][txIndexTo].PublicKey, 0)
	if reflect.DeepEqual(newTx.GenerateTransactionHash(mimc.NewMiMC()), []byte("")) {
		return &types.Transaction{}, errors.New("wrong tx hash")
	}
	(*repetitive)[s.AddressMap[shardID][txIndexFrom].Index] = append((*repetitive)[s.AddressMap[shardID][txIndexFrom].Index], s.AddressMap[indexTo][txIndexTo].Index)
	(*noncer)[s.AddressMap[shardID][txIndexFrom].Index] += 1
	(*counter)[s.AddressMap[shardID][txIndexFrom].Index] += 1
	newTx.Time = time.Now()
	return &newTx, nil
}

func send(url string, msg []byte) {
	buff := bytes.NewBuffer(msg)
	http.Post("http://"+url, "application/json", buff)
}

func containsString(item string, items []string) bool {
	for _, val := range items {
		if val == item {
			return true
		}
	}
	return false
}

func containsUint64(item uint64, items []uint64) bool {
	for _, val := range items {
		if val == item {
			return true
		}
	}
	return false
}
