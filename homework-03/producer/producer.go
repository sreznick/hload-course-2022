package producer

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"main/common"
	"main/producer/server_backend"
)

const SQL_DRIVER = "postgres"

func ProducerRoutine(c common.KafkaConfig, p common.PostgresConfig) {
	server_backend.SetProducerKafka(c)

	fmt.Println(sql.Drivers())
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		p.Host, p.Port, p.User, p.Password, p.Dbname)

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

	_, err = conn.Exec("create table if not exists urls(id bigint unique, url varchar unique, clicks int default 0)")
	if err != nil {
		fmt.Println("Failed to create table", err)
		panic("exit")
	}

	go server_backend.GetClicks(conn)
	r := server_backend.SetupRouter(conn)
	err = r.Run(":8080")
	if err != nil {
		panic("Something wrong with router: " + err.Error())
	}
}
