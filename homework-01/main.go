package main

import (
	"database/sql"
	"encoding/base32"
	"fmt"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"strconv"
	"time"
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

func GetIdByTinyUrl(tinyurl string, c *gin.Context) int {
	data, err := base32.StdEncoding.WithPadding(-1).DecodeString(tinyurl)
	if err != nil {
		c.Writer.WriteHeader(404)
		return -1
	}
	id, err := strconv.Atoi(string(data[:]))
	if err != nil {
		c.Writer.WriteHeader(404)
		return -1
	}
	return id
}

var (
	createRequestsNumber = promauto.NewCounter(prometheus.CounterOpts{
		Name: "Put_requests",
		Help: "Number of /create PUT requests",
	})

	createRequestTime = promauto.NewSummary(prometheus.SummaryOpts{
		Name: "Put_request_time",
		Help: "Time of /create PUT",
	})

	getRequestsNumber = promauto.NewCounter(prometheus.CounterOpts{
		Name: "Get_requests",
		Help: "Number of /:tinyurl GET requests",
	})

	getRequestTime = promauto.NewSummary(prometheus.SummaryOpts{
		Name: "Get_request_time",
		Help: "Time of GET",
	})
)

const (
	insertIntoUrls = "INSERT INTO urls (url) VALUES ($1) ON CONFLICT DO NOTHING"
	getIdByUrl     = "SELECT id FROM urls WHERE url = $1"
	getUrlById     = "SELECT url FROM urls WHERE id = $1"
	createTable    = "CREATE TABLE IF NOT EXISTS urls (id SERIAL, url TEXT PRIMARY KEY)"
)

func doCreate(c *gin.Context, db *sql.DB) {
	bodyRequest := Request{}
	err := c.BindJSON(&bodyRequest)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"response": "bad long url"})
		return
	}
	_, err = db.Exec(insertIntoUrls, bodyRequest.Long_url)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"response": "can't insert new long url: " + err.Error()})
		return
	}

	var id int
	err = db.QueryRow(getIdByUrl, bodyRequest.Long_url).Scan(&id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"response": "can't retrieve just inserted url id: " + err.Error()})
		return
	}

	tinyUrl := GetTinyUrlById(id)
	c.JSON(http.StatusOK, gin.H{"longurl": bodyRequest.Long_url, "tinyurl": tinyUrl})
}

func doUrlGet(c *gin.Context, db *sql.DB) {
	tinyUrl := c.Params.ByName("url")
	id := GetIdByTinyUrl(tinyUrl, c)
	if tinyUrl != GetTinyUrlById(id) {
		c.Writer.WriteHeader(404)
		return
	}
	var longUrl string
	err := db.QueryRow(getUrlById, id).Scan(&longUrl)
	if err != nil {
		c.Writer.WriteHeader(http.StatusNotFound)
	} else {
		c.Redirect(http.StatusFound, longUrl)
	}
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
		createRequestsNumber.Inc()
		start := time.Now()

		doCreate(c, db)

		elapsed := float64(time.Since(start).Milliseconds())
		createRequestTime.Observe(elapsed)
	})

	r.GET("/:url", func(c *gin.Context) {
		getRequestsNumber.Inc()
		start := time.Now()

		doUrlGet(c, db)

		elapsed := float64(time.Since(start).Milliseconds())
		getRequestTime.Observe(elapsed)
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

	_, err = conn.Exec(createTable)
	ErrorCheck(err, "ERROR: Failed to open or create table")

	return conn
}

func main() {
	conn := StartSQL()

	http.Handle("/metrics", promhttp.Handler())
	go func() {
		err := http.ListenAndServe(":2112", nil)
		ErrorCheck(err, "ERROR: Cannot connect to Prometheus")
	}()

	r := setupRouter(conn)
	r.Run(":8080")
}
