package main

import (
	"database/sql"
	"fmt"
	"net/http"

	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const SQL_DRIVER = "postgres"

const (
	host     = "localhost"
	port     = 5432
	user     = "gouser"
	password = "gouser"
	dbname   = "shorturl"
)

func main() {
	fmt.Println(sql.Drivers())

	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	conn, err := sql.Open(SQL_DRIVER, psqlconn)
	if err != nil {
		fmt.Println("Failed to open", err)
		panic("exit")
	}

	err = conn.Ping()
	if err != nil {
		fmt.Println("Failed to ping database", err)
		panic("exit")
	}

	createTable(conn)

	http.Handle("/metrics", promhttp.Handler())
	go func() { http.ListenAndServe(":2112", nil) }()

	r := setupRouter(conn)
	r.Run(":8080")
}
