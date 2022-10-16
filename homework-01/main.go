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

func main() {
	fmt.Println(sql.Drivers())
	conn := database.CreateConnection()
	conn.SetupTable()

	r := server.SetupRouter(conn)

	http.Handle("/metrics", promhttp.Handler())
	go func() {
		err := http.ListenAndServe(":8088", nil)
		if err != nil {
			panic("Problems with prometheus: " + err.Error())
		}
	}()

	if err := r.Run(":8080"); err != nil {
		fmt.Println("Failed to run ", err.Error())
		panic("exit")
	}
}
