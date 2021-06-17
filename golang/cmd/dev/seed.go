package main

import (
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

func main() {
	var (
		db  *sqlx.DB
		err error
	)

	for {
		//TODO: change to use local db package
		db, err = sqlx.Connect("mysql", "root@/")
		if err != nil {
			fmt.Println("waiting for mysql...")
			time.Sleep(time.Second)
			continue
		}
		break
	}
	for _, create := range structure {
		if _, err := db.Exec(create); err != nil {
			log.Fatal(err)
		}
	}
}

var createDatabase = `CREATE DATABASE jobcoin_mixer;`

var deposit_addresses = `CREATE TABLE jobcoin_mixer.deposit_addresses (
	deposit_address varchar(255),
	child varchar(255),
	UNIQUE KEY deposit_address_child(deposit_address, child)
);`

var mixes = `CREATE TABLE jobcoin_mixer.mixes (
	id int(11) AUTO_INCREMENT PRIMARY KEY,
	amount varchar(255),
	child varchar(255),
	status varchar(255),
	created_at datetime NOT NULL,
	updated_at datetime NOT NULL
);`

// TODO: this is unused, remove or refactor code / schema to reference
var add = `CREATE TABLE jobcoin_mixer.addresses (
	id int(11) AUTO_INCREMENT PRIMARY KEY,
	address varchar(255)
);`
var structure = []string{createDatabase, deposit_addresses, add}
