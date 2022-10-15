package main

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

const SQL_DRIVER = "postgres"
const (
	host     = "localhost"
	port     = 5432
	user     = "kirill"
	password = "111"
	dbname   = "kirill"
)

func setup() {
	db := setupModels()
	r := setupRouter(db)
	r.Run(":8081")
}

func setupModels() *sql.DB {
	psqlInfo := fmt.Sprintf("postgres://%v:%v@%v:%v/%v?sslmode=disable", user, password, host, port, dbname)
	conn, err := sql.Open(SQL_DRIVER, psqlInfo)
	if err != nil {
		fmt.Println("Failed to open", err)
		panic("exit")
	}

	err = conn.Ping()
	if err != nil {
		fmt.Println("Failed to ping database ", err)
		panic("exit")
	}

	_, err = conn.Exec("CREATE TABLE IF NOT EXISTS urls (id SERIAL, url TEXT PRIMARY KEY)")
	if err != nil {
		fmt.Println("Failed to create table", err)
		panic("exit")
	}
	return conn
}

func setupRouter(db *sql.DB) *gin.Engine {
	r := gin.Default()

	r.PUT("/create", func(c *gin.Context) {
		recordCreateStart()

		start := time.Now()

		handleCreate(c, db)

		recordCreateTime(float64(time.Since(start).Milliseconds()))
	})

	r.GET("/:url", func(c *gin.Context) {
		recordGetStart()

		start := time.Now()

		handleUrlGet(c, db)

		recordCreateTime(float64(time.Since(start).Milliseconds()))
	})

	return r
}

func main() {
	setup()
}
