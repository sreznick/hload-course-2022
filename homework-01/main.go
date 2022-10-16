package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"main/database"
	"main/server"
)

func main() {
	fmt.Println(sql.Drivers())
	conn := database.CreateConnection()
	conn.SetupTable()

	r := server.SetupRouter(conn)
	err := r.Run(":8080")
	if err != nil {
		fmt.Println("Failed to run", err.Error())
		panic("exit")
	}
}
