# Golang Jobcoin Mixer Implemention

## Overview of this document
This project is a working implementation of a Jobcoin mixer. It is not complete, does not include tests and does not exemplify what the final best version is. But, it works pretty well, and I had a bunch of fun making it!

I have left some "`TODO:...`" comments in the code intentionally to point out areas that I would revisit given more time. The remainder of the document outlines what I have added to the boilerplate in order to meet the minimum requirements. Finally, I have some instructions for how you can run this program locally, on a mac.

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
- instantiate and keep alive the `poller` package

The `POST /create` handler:
- creates a new Deposit Address that the requester can deposit funds into (much like the `mixer-cli`).
- stores a list of new "child" addresses (requester identified addresses for deposit) in the `jobcoin_mixer` database and associates them with the newly created Deposit Address
- registers the new deposit Address in the `Poller` service.


### Clean + Deps + Build
    `make all`

### Clean
    `make clean`

### Deps
    `make deps`

### Build
    `make build`

### Test
    `make test`

### Run
    `./bin/mixer --addresses=bravo,tango,delta`
