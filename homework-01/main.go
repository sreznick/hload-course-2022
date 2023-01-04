package main

import (
	"net/http"

	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	dbi "main/db_interactor"

	srv "main/server_api"
)

func main() {
	var db = dbi.OpenSQLConnection()
	db.CreateTableIfNotExists()

	http.Handle("/metrics", promhttp.Handler())
	go func() {
		err := http.ListenAndServe(":2112", nil)
		if err != nil {
			panic("Problems with prometheus: " + err.Error())
		}
	}()

	router := srv.SetupRouter(&db)

	router.Run("localhost:8080")
}
