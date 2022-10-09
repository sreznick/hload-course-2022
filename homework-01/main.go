package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

const SQL_DRIVER = "postgres"

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "jaja"
	dbname   = "hload"
)

type Body struct {
	Longurl string `json:"longurl"`
}

func setupRouter(db *sql.DB) *gin.Engine {
	r := gin.Default()

	r.PUT("/create", func(c *gin.Context) {
		body := Body{}
		err := c.BindJSON(&body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"response": "wrong json format"})
			return
		}

		_, err = db.Exec("insert into urls(url) values ($1) on conflict do nothing", body.Longurl)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"response": "something wrong with database: " + err.Error()})
			return
		}

		var shorturl int
		err = db.QueryRow("select id from urls where url = $1", body.Longurl).Scan(&shorturl)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"response": "something wrong with database: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"ans": shorturl})
	})

	r.GET("/:url", func(c *gin.Context) {
		shorturl := c.Params.ByName("url")

		// TEMPORARY
		id, _ := strconv.Atoi(shorturl)
		err := db.QueryRow("select url from urls where id = $1", id).Scan(&shorturl)
		if err != nil {
			c.Writer.WriteHeader(404)
			return
		}

		c.Redirect(302, shorturl)
	})

	return r
}

func main() {
	fmt.Println(sql.Drivers())
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

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

	_, err = conn.Exec("create table if not exists urls(id serial, url varchar unique)")
	if err != nil {
		fmt.Println("Failed to create table", err)
		panic("exit")
	}

	r := setupRouter(conn)
	r.Run(":8080")
}
