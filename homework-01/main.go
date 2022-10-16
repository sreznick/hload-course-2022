package main

import (
	"database/sql"
	"fmt"
	"net/http"

	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"main/server"
)

const SQL_DRIVER = "postgres"
const SQL_CONNECT_URL = "postgres://postgres:postgres@localhost"

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "persikk"
	dbname   = "hload1"
)

func checkError(err error, msg string) {
	if err != nil {
		fmt.Println(msg, err)
		panic("exit")
	}
}

func main() {
	fmt.Println(sql.Drivers())
	sqlParams := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	conn, err := sql.Open(SQL_DRIVER, sqlParams)
	checkError(err, "Failed to open")

	err = conn.Ping()
	checkError(err, "Failed to ping database")

	_, err = conn.Exec("create table if not exists urlsStorage(id serial primary key, long_url varchar(200));")
	checkError(err, "Failed to create urls table")

	http.Handle("/metrics", promhttp.Handler())
	go func() {
		err := http.ListenAndServe(":2112", nil)
		checkError(err, "Unable to connect Prometheus")
	}()

	r := server.SetupRouter(conn)
	r.Run(":8080")
}
