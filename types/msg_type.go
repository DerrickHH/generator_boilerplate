package types

type AccountsMsg struct {
	Content       [][]byte `json:"content"`
	AddressNumber int      `json:"number"`
}
