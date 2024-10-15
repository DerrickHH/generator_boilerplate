package types

import (
	"encoding/json"
	"github.com/consensys/gnark-crypto/ecc/bn254/fr"
	"github.com/consensys/gnark-crypto/ecc/bn254/twistededwards/eddsa"
	"hash"
	"time"
)

type Transaction struct {
	Nonce          uint64
	Amount         fr.Element
	SenderPubKey   eddsa.PublicKey
	ReceiverPubKey eddsa.PublicKey
	Signature      eddsa.Signature
	Time           time.Time
}

func NewTransaction(amount uint64, from, to eddsa.PublicKey, nonce uint64) Transaction {
	var t Transaction
	t.Amount = fr.NewElement(amount)
	t.SenderPubKey = from
	t.ReceiverPubKey = to
	t.Nonce = nonce
	return t
}

func (t *Transaction) SetSign(hFunc hash.Hash, privateKey eddsa.PrivateKey) {
	t.Signature = t.Sign(hFunc, privateKey)
}

func (t *Transaction) Sign(hFunc hash.Hash, privateKey eddsa.PrivateKey) eddsa.Signature {
	msg := t.GenerateTransactionHash(hFunc)
	return Sign(msg, privateKey, hFunc)
}

func (t *Transaction) GenerateTransactionHash(hFunc hash.Hash) []byte {
	hFunc.Reset()

	var frNonce fr.Element

	// conver uint64 to bytes
	frNonce.SetUint64(t.Nonce)
	nonceBytes := frNonce.Bytes()
	hFunc.Write(nonceBytes[:])
	buf1 := t.Amount.Bytes()
	hFunc.Write(buf1[:])

	buf1 = t.SenderPubKey.A.X.Bytes()
	hFunc.Write(buf1[:])

	buf1 = t.SenderPubKey.A.Y.Bytes()
	hFunc.Write(buf1[:])

	buf1 = t.ReceiverPubKey.A.X.Bytes()
	hFunc.Write(buf1[:])

	buf1 = t.ReceiverPubKey.A.Y.Bytes()
	hFunc.Write(buf1[:])

	hashSum := hFunc.Sum(nil)

	return hashSum
}

func (t *Transaction) VerifySignature(hFunc hash.Hash) (bool, error) {
	msg := t.GenerateTransactionHash(hFunc)
	return Verify(msg, t.SenderPubKey, t.Signature.Bytes(), hFunc)
}

func (t *Transaction) Marshal() ([]byte, error) {
	encoded, err := json.Marshal(t)
	if err != nil {
		return nil, err
	}
	return encoded, nil
}

func (t *Transaction) Unmarshal(content []byte) error {
	err := json.Unmarshal(content, t)
	if err != nil {
		return err
	}
	return nil
}
