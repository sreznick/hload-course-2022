package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"main/database"
	"main/server"
	"net/http"
)

const prometheusAddress = ":8088"
const appAddress = ":8080"

func main() {
	fmt.Println(sql.Drivers())
	conn := database.CreateConnection()
	conn.SetupTable()

	r := server.SetupRouter(conn)

	http.Handle("/metrics", promhttp.Handler())
	go func() {
		err := http.ListenAndServe(prometheusAddress, nil)
		if err != nil {
			panic("Problems with prometheus: " + err.Error())
		}
	}()

	if err := r.Run(appAddress); err != nil {
		fmt.Println("Failed to run ", err.Error())
		panic("exit")
	}
}
