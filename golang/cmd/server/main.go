package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/bitly/go-simplejson"
	"github.com/gemini/jobcoin/cmd/server/poller"
	"github.com/gemini/jobcoin/models/addresses"
	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
)

func wrapHandler(h http.Handler) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		// Take the context out from the request
		ctx := r.Context()

		// Get new context with key-value "params" -> "httprouter.Params"
		ctx = context.WithValue(ctx, "params", ps)

		// Get new http.Request with the new context
		r = r.WithContext(ctx)

		// Call your original http.Handler
		h.ServeHTTP(w, r)
	}
}

func handleRequests(p *poller.DepositAddressPoller) {
	port := ":1337"
	router := httprouter.New()
	// router.POST("/create", wrapHandler(CreateDepositAddress))
	router.POST("/create", CreateDepositAddress(p))
	fmt.Printf("\n=> listening at http://localhost%s\n", port)
	log.Fatal(http.ListenAndServe(port, router))

}

func main() {
	p, err := poller.InitializeDepositPoller()
	if err != nil {
		log.Fatal(err)
	}
	handleRequests(p)
}

// Caller provides jobcoin addresses they own
type NewAddress struct {
	Children []*addresses.ChildAddress `json:"deposit_addresses"`
}

func CreateDepositAddress(p *poller.DepositAddressPoller) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Fatalln(err)
		}

		childAddresses, err := simplejson.NewJson(body)
		if err != nil {
			log.Fatalln(err)
		}

		depositAddress, err := uuid.NewUUID()
		if err != nil {
			log.Fatal(err)
		}

		addresses.AddNewDepositAddress(depositAddress, childAddresses.MustStringArray())
		p.AddURL(addresses.NewDepositAddress(depositAddress.String()))
	}
}
