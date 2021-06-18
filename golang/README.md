# Golang Jobcoin Mixer Implemention

## Overview of this document
This project is a working implementation of a Jobcoin mixer. It is not complete, does not include tests and does not exemplify what the final best version is. But, it works pretty well, covers the base cases, and I had a bunch of fun making it!

- I have left some "`TODO:...`" comments in the code intentionally to point out areas that I would revisit given more time.
- The remainder of the document outlines what I have added to the boilerplate in order to meet the minimum requirements.
- There is a brief section on usage, with instructions for how you can run this program locally, (on a mac).
- Finally a brief discussion on what I'd change given more time.

## Poller
The poller package holds code to create and manage a polling service which "pings" the jobcoin server for positive balances in Deposit Addresses that are generated and maintained by this Jobcoin Mixer. New Deposit Addresses can be added dynamically, and if a positive balance on a Deposit Address is found on the Jobcoin network... _things will happen..._:

- Funds from the Deposit Address will be transfered to the "House Address"
- Transaction records will be added to the `jobcoin_mixer.mixes` table identifying the amounts that should be transfered FROM the House Address TO the Deposit Address' "children" (i.e. the addresses that were initially supplied to the Mixer)
- A function (that is a placeholder for another async "poller") is called that transfers funds from the House address to the child addresses (and marks the "mixes" complete upon success).
## db
The db package holds was designed to hold files relevant to db interaction. Moving that logic is left as a `TODO`. For now, it holds a function that returns a connection to the `jobcoin_mixer` mysql database for interaction.
## Server
I have added an HTTP server, that listens at port `:1337` on localhost. The server has two main functions:
- listen for `POST /create` requests
- instantiate and keep alive the `poller` package/service

The `POST /create` handler:
- creates a new Deposit Address that the requester can deposit funds into (much like the `mixer-cli`).
- stores a list of new "child" addresses (requester identified addresses for deposit) in the `jobcoin_mixer` database and associates them with the newly created Deposit Address.
- registers the new deposit Address in the `Poller` service.
- returns the new Deposit Address associated with the new child addresses in the response.

## Usage
prereq:
- get [homebrew](https://brew.sh/) if you don't have it.
- I used Go version 1.16.4 (via [asdf](https://asdf-vm.com/#/))

Get mysql:
```
$ brew install mysql@5.7 # that's the version I used for this
$ brew services start mysql@5.7 # if that ^ command doesn't start it for you (can check with `brew services list`)
```
Go get stuff:
```
$ go get
```

Seed the database:
```
$ go run cmd/dev/seed.go
```

Run the server:
```
$ go run server/main.go
```

Hit the `POST /create` endpoint with your addresses, e.g:
```
$ curl --header "Content-Type: application/json" \
  --request POST \
  --data '["alfa", "brafo", "foo", "bar"]' \
  http://localhost:1337/create

^ => returns new deposit address, e.g.: {"address":"0857920e-cff7-11eb-aa09-faffc20de92f"}
```
Funds added to this returned address will eventually be transfered to the addresses given in the `POST`.

## TODOs

- Add tests.
- I would have liked to organized things better in terms of models, database interaction, API. It was a small enough project that I was keeping things in the same files, but I'd definitely organize this better given more time
- Use a queue system to handle the transfers
- take a fee for transactions
- use a randomizer to divide the initial deposit amounts, and to schedule the time of transfer.
