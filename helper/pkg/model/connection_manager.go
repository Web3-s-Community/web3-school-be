package model

import (
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

func CreateConnection(host string, port int64, user, pass, dbName string) (*sqlx.DB, error) {
	db, err := sqlx.Connect("mysql", fmt.Sprintf("%v:%v@(%v:%v)/%v", user, pass, host, port, dbName))
	if err != nil {
		log.Fatalln(err)
	}
	return db, err
}

func CreateConnectionWithConnString(connString string) (*sqlx.DB, error) {
	db, err := sqlx.Connect("mysql", connString)
	if err != nil {
		log.Fatalln(err)
	}
	return db, err
}
