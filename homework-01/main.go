package main

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

const SQL_DRIVER = "postgres"
const SQL_CONNECT_URL = "postgres://postgres:postgres@localhost"
const SQL_Local = "host=localhost " +
	"port=5432 " +
	"user=postgres " +
	"password=psql " +
	"dbname=hload " +
	"sslmode=disable"

func main() {
	fmt.Println(sql.Drivers())

	conn, err := sql.Open(SQL_DRIVER, SQL_Local)
	if err != nil {
		fmt.Println("Failed to open", err)
		panic("exit")
	}

	err = conn.Ping()
	if err != nil {
		fmt.Println("Failed to ping database", err)
		panic("exit")
	}

	r := SetupRouter(conn)
	r.Run(":8080")
}
