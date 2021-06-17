package poller

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/gemini/jobcoin/cmd/server/mixer"
	"github.com/gemini/jobcoin/db"
	"github.com/gemini/jobcoin/models/addresses"
)

type DepositAddressPoller struct {
	ticker    *time.Ticker
	add       chan addresses.DepositAddress
	addresses []addresses.DepositAddress
}

func NewDepositPoller() *DepositAddressPoller {
	rv := &DepositAddressPoller{
		ticker: time.NewTicker(time.Second * 10),
		add:    make(chan addresses.DepositAddress),
	}
	go rv.run()
	return rv
}

func InitializeDepositPoller() (*DepositAddressPoller, error) {
	poller := NewDepositPoller()
	if err := registerDepositAddresses(poller); err != nil {
		return poller, err
	}
	return poller, nil
}

func registerDepositAddresses(h *DepositAddressPoller) error {
	depositAddresses, err := GetDepositAddresses()
	if err != nil {
		return err
	}

	for _, address := range depositAddresses {
		h.AddURL(address)
	}
	return nil
}

func (h *DepositAddressPoller) run() {
	client := &http.Client{}
	for {
		select {
		case <-h.ticker.C:
			for _, address := range h.addresses {
				processDepositAddresses(&address, client)
			}
		case address := <-h.add:
			h.addresses = append(h.addresses, address)
		}
	}
}

func (h *DepositAddressPoller) AddURL(address *addresses.DepositAddress) {
	fmt.Println("new deposit address: ", address)
	h.add <- *address
}
func GetDepositAddresses() ([]*addresses.DepositAddress, error) {
	var addrs []*addresses.DepositAddress
	conn := db.DBConn()
	defer conn.Close()
	rows, err := conn.Query("SELECT DISTINCT(deposit_address) from deposit_addresses")
	if err != nil {
		return addrs, err
	}
	var address string
	for rows.Next() {
		rows.Scan(&address)
		addrs = append(addrs, addresses.NewDepositAddress(address))
	}
	return addrs, nil
}

func processDepositAddresses(address *addresses.DepositAddress, client *http.Client) error {
	req, err := http.NewRequest("GET", address.URL(), nil)
	if err != nil {
		return err
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var addressInfo addresses.AddressResponse
	json.Unmarshal(bodyBytes, &addressInfo)
	log.Println("heartbeat ( address:", address.Address, ")")
	if addressInfo.BalanceFloat() > 0 {
		fmt.Println("Mixing ", address.Address, "(amount): ", addressInfo.Balance)
		if err := mixer.Mix(address, addressInfo); err != nil {
			return err
		}
	}

	return nil
}
