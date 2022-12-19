package main

import (
	"database/sql"
	"fmt"
	"net/http"

	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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

	_, err = conn.Exec("CREATE TABLE IF NOT EXISTS urlsStorage(id serial primary key, long_url varchar(200));")
	if err != nil {
		fmt.Println("Failed to create urls table", err)
		panic("exit")
	}

	http.Handle("/metrics", promhttp.Handler())
	go func() {
		err := http.ListenAndServe(":2112", nil)
		if err != nil {
			panic("Problems with prometheus: " + err.Error())
		}
	}()

	r := SetupRouter(conn)
	r.Run(":8080")
}
