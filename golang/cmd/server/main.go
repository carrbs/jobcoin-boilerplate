package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gemini/jobcoin/cmd/server/poller"
)

func handleRequests() {
	port := ":1337"
	fmt.Printf("\n=> listening at http://localhost%s\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

func main() {
	poller.InitializeDepositPoller()
	handleRequests()
}
