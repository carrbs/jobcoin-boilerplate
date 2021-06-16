package poller

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/gemini/jobcoin"
	"github.com/gemini/jobcoin/db"
)

type DepositAddressPoller struct {
	ticker    *time.Ticker
	add       chan DepositAddress
	addresses []DepositAddress
}

type DepositAddress struct {
	Address string
}

func (a *DepositAddress) URL() string {
	return jobcoin.AddressesEndpoint + "/" + a.Address
}

func NewDepositPoller() *DepositAddressPoller {
	rv := &DepositAddressPoller{
		ticker: time.NewTicker(time.Second * 1),
		add:    make(chan DepositAddress),
	}
	go rv.run()
	return rv
}

func InitializeDepositPoller() error {
	poller := NewDepositPoller()
	if err := registerDepositAddresses(poller); err != nil {
		return err
	}
	return nil
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
				url := address.URL()
				processDepositAddresses(url, client)
			}
		case address := <-h.add:
			h.addresses = append(h.addresses, address)
		}
	}
}

func (h *DepositAddressPoller) AddURL(address *DepositAddress) {
	fmt.Println("new deposit address: ", address)
	h.add <- *address
}
func GetDepositAddresses() ([]*DepositAddress, error) {
	var addresses []*DepositAddress
	conn := db.DBConn()
	defer conn.Close()
	rows, err := conn.Query("SELECT DISTINCT(deposit_address) from deposit_addresses")
	if err != nil {
		return addresses, err
	}
	var address string
	for rows.Next() {
		rows.Scan(&address)
		addresses = append(addresses, NewDepositAddress(address))
	}
	return addresses, nil
}

func NewDepositAddress(address string) *DepositAddress {
	return &DepositAddress{Address: address}
}

func processDepositAddresses(url string, client *http.Client) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Print(err.Error())
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		fmt.Print(err.Error())
	}
	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Print(err.Error())
	}
	var addressInfo AddressResponse
	json.Unmarshal(bodyBytes, &addressInfo)
	fmt.Println("bodyBytes", string(bodyBytes))

	fmt.Printf("API Response as struct %+v\n", addressInfo)

	if addressInfo.balance() > 0 {
		// Mix()
		fmt.Println("ALERT! Positive BALANCE!", addressInfo.Balance)
	}

	return nil
}

type Transaction struct {
	Timestamp   time.Time `json:"timestamp"`
	ToAddress   string    `json:"toAddress"`
	FromAddress string    `json:"fromAddress"`
	Amount      string    `json:"amount"`
}

type AddressResponse struct {
	Balance      string        `json:"balance"`
	Transactions []Transaction `json:"transactions"`
}

func (a *AddressResponse) balance() float64 {
	// FIXME: handle error
	f, _ := strconv.ParseFloat(a.Balance, 64)
	return f
}
