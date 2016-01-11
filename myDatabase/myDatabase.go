package myDatabase

import (
	"database/sql"
	"fmt"
	"log"

	"myMessages"

	_ "github.com/go-sql-driver/mysql"
)

func DBConnect() *sql.DB {

	db, err := sql.Open("mysql", "username:password@tcp(ip:port)/databaseName")
	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Println(myMessages.DBConnected)
	}

	//defer db.Close()
	return db
}
