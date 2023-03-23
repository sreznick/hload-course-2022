package main

import (
	"database/sql"
	"encoding/base32"
	"fmt"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"net/http"
	"strconv"
)

const SQL_DRIVER = "postgres"

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "postgres"
	dbname   = "hload_1"
)

type Request struct {
	Long_url string `json:"longurl"`
}

func GetTinyUrlById(id int) string {
	data := []byte(strconv.Itoa(id))
	str := base32.StdEncoding.WithPadding(-1).EncodeToString(data)
	return str
}

func GetIdByTinyUrl(tinyurl string) int {
	data, err := base32.StdEncoding.WithPadding(-1).DecodeString(tinyurl)
	ErrorCheck(err, "Decode error:")
	fmt.Println(tinyurl)
	fmt.Println(string(data[:]))
	fmt.Println(tinyurl)
	id, err := strconv.Atoi(string(data[:]))
	ErrorCheck(err, "Decode error while converting:")
	return id
}

func setupRouter(db *sql.DB) *gin.Engine {
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

	r.PUT("/create", func(c *gin.Context) {
		body_request := Request{}
		err := c.BindJSON(&body_request)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"response": "bad long url"})
			return
		}
		_, err = db.Exec("INSERT INTO urls (url) VALUES ($1) ON CONFLICT DO NOTHING", body_request.Long_url)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"response": "can't insert new long url: " + err.Error()})
			return
		}

		var id int
		err = db.QueryRow("SELECT id FROM urls WHERE url = $1", body_request.Long_url).Scan(&id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"response": "can't retrieve just inserted url id: " + err.Error()})
			return
		}

		tinyUrl := GetTinyUrlById(id)
		c.JSON(http.StatusOK, gin.H{"longurl": body_request.Long_url, "tinyurl": tinyUrl})
	})

	r.GET("/:url", func(c *gin.Context) {
		tinyUrl := c.Params.ByName("url")
		id := GetIdByTinyUrl(tinyUrl)
		var longUrl string
		err := db.QueryRow("SELECT url FROM urls WHERE id = $1", id).Scan(&longUrl)
		if err != nil {
			c.Writer.WriteHeader(http.StatusNotFound)
		} else {
			c.Redirect(http.StatusFound, longUrl)
		}
	})

	return r
}

func ErrorCheck(err error, message string) {
	if err != nil {
		fmt.Println(message, err)
		panic("exit")
	}
}

func StartSQL() *sql.DB {
	fmt.Println(sql.Drivers())
	sqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+"password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	conn, err := sql.Open(SQL_DRIVER, sqlInfo)
	ErrorCheck(err, "ERROR: Failed to open")

	err = conn.Ping()
	ErrorCheck(err, "ERROR: Failed to ping database")

	_, err = conn.Exec("CREATE TABLE IF NOT EXISTS urls (id SERIAL, url TEXT PRIMARY KEY)")
	ErrorCheck(err, "ERROR: Failed to open or create table")

	return conn
}

func main() {
	conn := StartSQL()
	r := setupRouter(conn)
	r.Run(":8080")
}
