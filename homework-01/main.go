package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"main/server_backend"
	"net/http"
)

const SQL_DRIVER = "postgres"

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "jaja"
	dbname   = "hload"
)

//TODO decompose
func main() {
	fmt.Println(sql.Drivers())
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	conn, err := sql.Open(SQL_DRIVER, psqlInfo)
	if err != nil {
		fmt.Println("Failed to open", err)
		panic("exit")
	}

	err = conn.Ping()
	if err != nil {
		fmt.Println("Failed to ping database", err)
		panic("exit")
	}

	_, err = conn.Exec("create table if not exists urls(id serial, url varchar unique)")
	if err != nil {
		fmt.Println("Failed to create table", err)
		panic("exit")
	}

	http.Handle("/metrics", promhttp.Handler())
	go func() {
		err := http.ListenAndServe(":2112", nil)
		if err != nil {
			panic("Problems with prometheus: " + err.Error())
		}
	}()

	r := server_backend.SetupRouter(conn)
	err = r.Run(":8080")
	if err != nil {
		panic("Something wrong with router: " + err.Error())
	}
}
