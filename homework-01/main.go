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

const (
	prometheusPort = ":2112"
	routerPort     = ":8080"
)

func main() {
	fmt.Println(sql.Drivers())

	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	db := getDatabaseInstance()
	db.InitDB(SQL_DRIVER, psqlconn)

	http.Handle("/metrics", promhttp.Handler())
	go func() { http.ListenAndServe(prometheusPort, nil) }()

	r := setupRouter()
	r.Run(routerPort)
}
