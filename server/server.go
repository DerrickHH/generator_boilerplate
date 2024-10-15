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
	"math/big"
	"net/http"
	"time"
)

var SequenceID = 1

type Server struct {
	Url string
	// NOTE: 用于生成transaction
	AddressMap map[int][]*types.AccountState
	// NOTE: 用于记录每个 shard 的 #0 节点
	ShardsTable map[string]string
	BeaconNode  string

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
		AddressMap:      make(map[int][]*types.AccountState),
		ShardsTable:     make(map[string]string),
		BeaconNode:      constant.BeaconNode,
		AccountsMsg:     make([]*types.AccountsMsg, 0),
		TransactionsMsg: make([]*types.RequestMsg, 0),

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
	accounts, err := generator.GenerateAccounts(msg.Number)
	if err != nil {
		return err
	}
	accountsMsg := &types.AccountsMsg{
		Content: make([][]byte, 0),
		ShardID: msg.ShardID,
	}
	s.AddressMap[msg.ShardID] = make([]*types.AccountState, 0)
	for _, acc := range accounts {
		accMar := acc.Marshal()
		accountsMsg.Content = append(accountsMsg.Content, accMar)
		s.AddressMap[msg.ShardID] = append(s.AddressMap[msg.ShardID], &acc)
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
	counter := make(map[string]int)
	// NOTE: 控制交易重复
	repetitive := make(map[string][]string)
	// NOTE: 控制 nonce
	noncer := make(map[string]int64)
	for _, acc := range s.AddressMap[req.ShardID] {
		counter[acc.Address] = 0
		repetitive[acc.Address] = make([]string, 0)
	}
	trans, ctrans := 0, 0
	for trans+ctrans < number {
		rnd, _ := rand.Int(rand.Reader, big.NewInt(100))
		if int(rnd.Int64()) > req.CrossShardRatio {
			tx, err := generator.GenerateTransaction(s.AddressMap[req.ShardID], &counter, &repetitive, &noncer)
			if err != nil {
				log.Println("[ERROR] Wrong when generating the transactions: ", err)
				continue
			}
			generatedTransactions = append(generatedTransactions, *tx)
			trans += 1
		} else {
			ctx, err := generator.GenerateCrossShardTransaction(req.ShardID, s.AddressMap, &counter, &repetitive, &noncer)
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

func send(url string, msg []byte) {
	buff := bytes.NewBuffer(msg)
	http.Post("http://"+url, "application/json", buff)
}
