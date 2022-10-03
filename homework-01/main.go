package main

import (
    "database/sql"
    "fmt"
    "net/http"

    "github.com/gin-gonic/gin"
    _ "github.com/lib/pq"
)

const SQL_DRIVER = "postgres"
const SQL_CONNECT_URL = "postgres://postgres:postgres@localhost"

func setupRouter() *gin.Engine {
	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	r.GET("/user/:name", func(c *gin.Context) {
		user := c.Params.ByName("name")
		if user == "vasya" {
			c.JSON(http.StatusOK, gin.H{"user": user, "value": "12345"})
		} else {
			c.JSON(http.StatusOK, gin.H{"user": user, "status": "no value"})
		}
	})

	return r
}

func main() {
    fmt.Println(sql.Drivers())
    conn, err := sql.Open(SQL_DRIVER, SQL_CONNECT_URL)
    if err != nil {
        fmt.Println("Failed to open", err)
        panic("exit")
    }

    err = conn.Ping()
    if err != nil {
        fmt.Println("Failed to ping database", err)
        panic("exit")
    }


	r := setupRouter()
	r.Run(":8080")
}
