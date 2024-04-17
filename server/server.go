package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"generator_boilerplate/constant"
	"generator_boilerplate/generator"
	"generator_boilerplate/types"
	"log"
	"net/http"
	"strconv"
)

type Server struct {
	Port string
	// NOTE: 用于生成transaction
	AddressMap map[int][]string
	// NOTE: 用于记录每个 shard 的 #0 节点
	ShardsTable map[string]string
}

func NewServer(port string) *Server {
	server := &Server{
		Port:        port,
		AddressMap:  make(map[int][]string),
		ShardsTable: make(map[string]string),
	}
	server.ShardsTable = constant.ShardsTable
	return server
}

func (s *Server) setRoutes() {
	http.HandleFunc("/generate_account", s.handleGenerateAccounts)
	// http.HandleFunc("/generate_transaction", s.handleGenerateTransaction)
}

func (s *Server) handleGenerateAccounts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowd", http.StatusMethodNotAllowed)
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

func (s *Server) Start() {
	s.setRoutes()
	fmt.Printf("Server is running on http://0.0.0.0:%s/\n", s.Port)
	err := http.ListenAndServe("0.0.0.0:"+s.Port, nil)
	if err != nil {
		log.Fatal(err)
	}
}
