package generator

import (
	"generator_boilerplate/types"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
)

func GenerateAccounts(number int) ([]*types.AccountState, map[uint64]types.SignatureAccount) {
	accounts := make([]*types.AccountState, number)
	accountsMap := make(map[uint64]types.SignatureAccount)
	for i := uint64(0); i < uint64(number); i++ {
		privKey, pubKey := types.GenerateKeys(int64(i))
		accountsMap[i] = types.SignatureAccount{
			PubKey:  pubKey,
			PrivKey: privKey,
		}
		chainAccount := types.AccountState{
			Index:     i,
			Nonce:     0,
			Balance:   fr.NewElement((i + 1) * 60),
			PublicKey: pubKey,
		}

		accounts = append(accounts, &chainAccount)
	}
	return accounts, accountsMap
}
