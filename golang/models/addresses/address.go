package addresses

import (
	"strconv"
	"time"

	"github.com/gemini/jobcoin"
	"github.com/gemini/jobcoin/db"
	"github.com/google/uuid"
)

type BaseAddress struct {
	Address string
}

type ChildAddress BaseAddress   // Where the mixer deposits to
type DepositAddress BaseAddress // Where people deposit

func (a *DepositAddress) URL() string {
	return jobcoin.AddressesEndpoint + "/" + a.Address
}

type AddressResponse struct {
	Balance      string        `json:"balance"`
	Transactions []Transaction `json:"transactions"`
}

func (a *AddressResponse) BalanceFloat() float64 {
	// FIXME: handle error
	f, _ := strconv.ParseFloat(a.Balance, 64)
	return f
}
func NewDepositAddress(address string) *DepositAddress {
	return &DepositAddress{Address: address}
}
func NewChildAddress(address string) *ChildAddress {
	return &ChildAddress{Address: address}
}

type Transaction struct {
	Timestamp   time.Time `json:"timestamp"`
	ToAddress   string    `json:"toAddress"`
	FromAddress string    `json:"fromAddress"`
	Amount      string    `json:"amount"`
}

func AddNewDepositAddress(depositAddress uuid.UUID, addresses []string) error {
	conn := db.DBConn()
	defer conn.Close()
	for _, child := range addresses {
		_, err := conn.Exec("INSERT INTO deposit_addresses(deposit_address, child) values (?,?)", depositAddress.String(), child)
		if err != nil {
			return err
		}
	}
	return nil
}
