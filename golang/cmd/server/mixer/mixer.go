package mixer

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gemini/jobcoin"
	"github.com/gemini/jobcoin/db"
	"github.com/gemini/jobcoin/models/addresses"
)

func Mix(address *addresses.DepositAddress, addressInfo addresses.AddressResponse) error {

	// 1. transfer full balance from depositAddress to HouseAddress
	if err := transferDepositToHouse(address, addressInfo); err != nil {
		return err
	}

	// 2. need to add to audit log/ledger
	if err := recordHouseTransfer(address, addressInfo); err != nil {
		return err
	}
	// 3. divvy up and transfer amongst deposit's children less fee
	// TODO: I would change this to be more async, possibly using something
	//       similar to the poller
	if err := processChildDeposits(); err != nil {
		return err
	}

	return nil
}

type Transfer struct {
	ToAddress   string `json:"toAddress"`
	FromAddress string `json:"fromAddress"`
	Amount      string `json:"amount"`
}

func processChildDeposits() error {
	addrs, err := getPendingChildDeposits()
	if err != nil {
		return err
	}

	transferHouseToChild(addrs)
	if err != nil {
		return err
	}
	return nil
}

func getPendingChildDeposits() ([]*MixDeposit, error) {
	conn := db.DBConn()
	defer conn.Close()
	var deposits []*MixDeposit
	rows, err := conn.Query("SELECT id, amount, child, status FROM mixes where status = ?", Pending)
	if err != nil {
		return deposits, err
	}
	for rows.Next() {
		var deposit MixDeposit
		rows.Scan(&deposit.Id, &deposit.Amount, &deposit.Child, &deposit.Status)
		deposits = append(deposits, &deposit)

	}
	return deposits, nil
}

func recordHouseTransfer(address *addresses.DepositAddress, addressInfo addresses.AddressResponse) error {
	addrs, err := getChildAddresses(address)
	if err != nil {
		return err
	}
	deposits := getChildDepositAmounts(addrs, addressInfo.BalanceFloat())
	conn := db.DBConn()
	defer conn.Close()
	now := time.Now()
	for childAddress, amount := range deposits {
		_, err := conn.Exec("INSERT INTO mixes (amount, child, status, created_at, updated_at) VALUES (?,?,?,?,?)", amount, childAddress, Pending, now, now)
		if err != nil {
			return err
		}

	}
	return nil
}

type childDeposits map[string]string

// Mirror database structure from `mixes` table
type MixDeposit struct {
	Id     int       `db:"id"`
	Amount string    `db:"amount"`
	Child  string    `db:"child"`
	Status MixStatus `db:"status"`
}

type MixStatus string

var (
	Pending  MixStatus = "pending"
	Complete MixStatus = "complete"
)

func getChildDepositAmounts(children []*addresses.ChildAddress, total float64) childDeposits {
	// TODO: make division a random amount, make this generally more interesting
	deposits := make(childDeposits)
	partition := total / float64(len(children))
	deposit := strconv.FormatFloat(partition, 'g', -1, 64)
	for _, child := range children {
		deposits[child.Address] = deposit
	}
	return deposits

}

func getChildAddresses(address *addresses.DepositAddress) ([]*addresses.ChildAddress, error) {
	var addrs []*addresses.ChildAddress
	conn := db.DBConn()
	defer conn.Close()
	rows, err := conn.Query("SELECT child FROM deposit_addresses where deposit_address = ? ", address.Address)
	if err != nil {
		return addrs, err
	}
	var a string
	for rows.Next() {

		rows.Scan(&a)
		addrs = append(addrs, addresses.NewChildAddress(a))
	}
	return addrs, nil
}

func transferDepositToHouse(address *addresses.DepositAddress, addressInfo addresses.AddressResponse) error {
	transferPayload, err := json.Marshal(&Transfer{
		ToAddress:   jobcoin.HouseAccountAddress,
		FromAddress: address.Address,
		Amount:      addressInfo.Balance,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", jobcoin.TransactionEndpoint, bytes.NewBuffer(transferPayload))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json; charset=UTF-8")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func transferHouseToChild(deposits []*MixDeposit) error {
	// TODO: subtract/transfer a jobcoin as processing fee
	for _, deposit := range deposits {
		transferPayload, err := json.Marshal(&Transfer{
			ToAddress:   deposit.Child,
			FromAddress: jobcoin.HouseAccountAddress,
			Amount:      deposit.Amount,
		})
		if err != nil {
			return err
		}
		req, err := http.NewRequest("POST", jobcoin.TransactionEndpoint, bytes.NewBuffer(transferPayload))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json; charset=UTF-8")
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			if err := markMixComplete(deposit); err != nil {
				return err
			}
		}
	}
	return nil
}
func markMixComplete(deposit *MixDeposit) error {
	conn := db.DBConn()
	defer conn.Close()
	now := time.Now()
	_, err := conn.Exec("UPDATE mixes SET status = ?, updated_at = ? where id = ?", Complete, now, deposit.Id)
	if err != nil {
		return err
	}

	return nil
}
