package generator

import (
	"fmt"
	"generator_boilerplate/constant"
	"generator_boilerplate/types"
	"github.com/ethereum/go-ethereum/crypto"
	"log"
)

func GenerateAccounts(number int) ([]types.Account, error) {
	var accounts []types.Account
	for i := 0; i < number; i++ {
		privateKey, err := crypto.GenerateKey()
		if err != nil {
			log.Fatalf("Failed to generate private key: %v", err)
		}

		privateKeyBytes := crypto.FromECDSA(privateKey)
		address := crypto.PubkeyToAddress(privateKey.PublicKey).Hex()

		accounts[i] = types.Account{
			PrivateKey: fmt.Sprintf("0x%x", privateKeyBytes),
			Address:    address,
			Balance:    constant.Balance,
			Nonce:      0,
			ShardList:  make([]int, 0),
		}
	}
	return accounts, nil
}
