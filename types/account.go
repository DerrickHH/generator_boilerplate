package types

import (
	"encoding/binary"
	"fmt"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark-crypto/ecc/bn254/twistededwards/eddsa"
)

type Address = string

const AccountSizeInBytes = 160

type Signature struct {
	PubKey  eddsa.PublicKey
	PrivKey eddsa.PrivateKey
}

type AccountState struct {
	Index     uint64
	Address   Address
	PublicKey eddsa.PublicKey
	Nonce     uint64
	Balance   fr.Element
}

func (as *AccountState) Reset() {
	as.Index = 0
	as.Nonce = 0
	as.Balance.SetZero()
	as.PublicKey.A.X.SetZero()
	as.PublicKey.A.Y.SetZero()
}

func (as *AccountState) Marshal() []byte {
	res := [AccountSizeInBytes]byte{}
	binary.BigEndian.PutUint64(res[24:], as.Index)
	binary.BigEndian.PutUint64(res[56:], as.Nonce)
	buf := as.Balance.Bytes()
	copy(res[64:], buf[:])
	buf = as.PublicKey.A.X.Bytes()
	copy(res[96:], buf[:])
	buf = as.PublicKey.A.Y.Bytes()
	copy(res[128:], buf[:])
	return res[:]
}

func UnMarshal(acc *AccountState, accBytes []byte) error {
	if len(accBytes) != AccountSizeInBytes {
		return fmt.Errorf("invalid bytes: required %d bytes, but found %d bytes", AccountSizeInBytes, len(accBytes))
	}
	acc.Index = binary.BigEndian.Uint64(accBytes[24:32])
	acc.Nonce = binary.BigEndian.Uint64(accBytes[56:64])
	acc.Balance.SetBytes(accBytes[64:96])
	acc.PublicKey.A.X.SetBytes(accBytes[96:128])
	acc.PublicKey.A.Y.SetBytes(accBytes[128:])
	return nil
}
