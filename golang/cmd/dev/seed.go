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
		// db, err = sqlx.Connect("mysql", "root@/?timeout=30s&readTimeout=30s&writeTimeout=30s&multiStatements=true&charset=utf8mb4")
		db, err = sqlx.Connect("mysql", "root@/")
		if err != nil {
			fmt.Println("err: ", err)
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

var add = `CREATE TABLE jobcoin_mixer.addresses (
	id int(11) AUTO_INCREMENT PRIMARY KEY,
	address varchar(255)
);`
var structure = []string{createDatabase, deposit_addresses, add}
