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

const (
	dbCreate = "create table if not exists urlsStorage(id serial primary key, long_url varchar(200));"

	host     = "localhost"
	port     = 5432
	user     = "mandelshtamd"
	password = "admin"
	dbname   = "tinyurls"

	serverPortNumber     = ":8080"
	prometheusPortNumber = ":2112"
)

func main() {
	fmt.Println(sql.Drivers())
	sqlParams := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	conn, err := sql.Open(SQL_DRIVER, sqlParams)
	server.HandleError(err, "Unable to connect db")
	err = conn.Ping()
	_, err = conn.Exec(dbCreate)

	http.Handle("/metrics", promhttp.Handler())
	go func() {
		err := http.ListenAndServe(prometheusPortNumber, nil)
		server.HandleError(err, "Unable to connect prometheus")
	}()

	r := server.Setup(conn)
	r.Run(serverPortNumber)
}
