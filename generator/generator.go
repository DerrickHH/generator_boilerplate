package generator

import (
	"crypto/rand"
	"errors"
	"fmt"
	"generator_boilerplate/constant"
	"generator_boilerplate/types"
	"github.com/ethereum/go-ethereum/crypto"
	"log"
	"math/big"
	"reflect"
)

func GenerateAccounts(number int) ([]types.AccountState, error) {
	accounts := make([]types.AccountState, number)
	for i := 0; i < number; i++ {
		privateKey, err := crypto.GenerateKey()
		if err != nil {
			log.Fatalf("Failed to generate private key: %v", err)
		}

		// privateKeyBytes := crypto.FromECDSA(privateKey)
		address := crypto.PubkeyToAddress(privateKey.PublicKey).Hex()

		accounts[i] = types.AccountState{
			PublicKey: crypto.FromECDSAPub(&privateKey.PublicKey),
			Address:   address,
			Balance:   big.NewInt(constant.Balance),
			Nonce:     0,
			// ShardList:  make([]int, 0),
		}
	}
	return accounts, nil
}

func GenerateTransaction(addresses []types.Account, counter *map[string]int, repetitive *map[string][]string, noncer *map[string]int64) (*types.Transaction, error) {
	indexFrom, indexTo := 0, 0
	for indexFrom == indexTo {
		indexFromInt64, _ := rand.Int(rand.Reader, big.NewInt(int64(len(addresses))))
		indexFrom = int(indexFromInt64.Int64())
		indexToInt64, _ := rand.Int(rand.Reader, big.NewInt(int64(len(addresses))))
		indexTo = int(indexToInt64.Int64())
	}
	if containsString(addresses[indexTo].Address, (*repetitive)[addresses[indexFrom].Address]) {
		return &types.Transaction{}, errors.New("repetitive from and to")
	}
	if (*counter)[addresses[indexFrom].Address] >= constant.MaxTxsInBlock {
		return &types.Transaction{}, errors.New("transaction counter has exceed")
	}
	if addresses[indexFrom].Balance < 1 {
		return &types.Transaction{}, errors.New("no sufficient balance")
	}
	newTx := types.NewTransaction(addresses[indexFrom].Address,
		addresses[indexTo].Address, 1, (*noncer)[addresses[indexFrom].Address])
	_ = newTx.GenerateTransactionHash()
	if reflect.DeepEqual(newTx.Hash, []byte("")) {
		// log.Error("[ERROR] Wrong when generate the transaction: nil hash.")
		return &types.Transaction{}, errors.New("wrong tx hash")
	}
	(*repetitive)[addresses[indexFrom].Address] = append((*repetitive)[addresses[indexFrom].Address], addresses[indexTo].Address)
	(*noncer)[addresses[indexFrom].Address] += 1
	(*counter)[addresses[indexFrom].Address] += 1
	return &newTx, nil
}

func GenerateCrossShardTransaction(shardID int, addressMap map[int][]types.Account, counter *map[string]int, repetitive *map[string][]string, noncer *map[string]int64) (*types.CrossShardTransaction, error) {
	fmt.Println("The length of addressMap is: ", len(addressMap))
	txIndexFromInt64, _ := rand.Int(rand.Reader, big.NewInt(int64(len(addressMap[shardID]))))
	txIndexFrom := int(txIndexFromInt64.Int64())
	if (*counter)[addressMap[shardID][txIndexFrom].Address] >= constant.MaxTxsInBlock {
		return &types.CrossShardTransaction{}, errors.New("counter has exceed")
	}
	if addressMap[shardID][txIndexFrom].Balance < 1 {
		return &types.CrossShardTransaction{}, errors.New("no sufficient balance")
	}
	indexTo := shardID
	for indexTo == shardID {
		// 根据 全局 的 constant 计算 目标 shard 是谁
		indexToInt64, _ := rand.Int(rand.Reader, big.NewInt(int64(len(addressMap))))
		indexTo = int(indexToInt64.Int64())
	}
	fmt.Println("indexTo is: ", indexTo)
	txIndexToInt64, _ := rand.Int(rand.Reader, big.NewInt(int64(len(addressMap[indexTo]))))
	txIndexTo := int(txIndexToInt64.Int64())
	// 如果这对组合的交易已经存在的，也不能保留
	if containsString(addressMap[indexTo][txIndexTo].Address, (*repetitive)[addressMap[shardID][txIndexFrom].Address]) {
		return &types.CrossShardTransaction{}, errors.New("repetitive from and to")
	}
	newTx := types.NewCrossShardTransaction(shardID, addressMap[shardID][txIndexFrom].Address,
		addressMap[indexTo][txIndexTo].Address, 1, (*noncer)[addressMap[shardID][txIndexFrom].Address])
	_ = newTx.GenerateTransactionHash()
	if reflect.DeepEqual(newTx.Hash, []byte("")) {
		return &types.CrossShardTransaction{}, errors.New("wrong tx hash")
	}
	(*repetitive)[addressMap[shardID][txIndexFrom].Address] = append((*repetitive)[addressMap[shardID][txIndexFrom].Address], addressMap[indexTo][txIndexTo].Address)
	(*noncer)[addressMap[shardID][txIndexFrom].Address] += 1
	(*counter)[addressMap[shardID][txIndexFrom].Address] += 1
	return &newTx, nil
}

func containsString(item string, items []string) bool {
	for _, val := range items {
		if val == item {
			return true
		}
	}
	return false
}
