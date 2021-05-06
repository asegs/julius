package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
)

func openConnection() *sql.DB{
	fmt.Println("Go MySQL Connection Initiated")

	db,err := sql.Open("mysql","root:@tcp(127.0.0.1:3306)/livechat")
	if err != nil{
		panic(err.Error())
	}

	return db
}