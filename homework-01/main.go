package main

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	SQL_DRIVER      = "postgres"
	SQL_CONNECT_URL = "postgres://postgres:postgres@localhost"
)

type DBNotFoundError struct {
	what string
}

func (e *DBNotFoundError) Error() string {
    return "\"" + e.what + "\" not found."
}

func New(text string) *DBNotFoundError {
	return &DBNotFoundError{text}
}

func DBGetLong(db *sql.DB, shortURL string) (string, error) {
	if shortURL == "" {
		return "https://youtu.be/dQw4w9WgXcQ", nil
	}
	rows, err := db.Query(`SELECT longs FROM URLs WHERE shorts=$1`, shortURL)
	if err != nil {
		return "", err
	}

	for rows.Next() {
		var longURL string
		err = rows.Scan(&longURL)
		return longURL, err
	}
	return "", New("shortURL")
}

func DBGetShort(db *sql.DB, longURL string) (string, error) {
	rows, err := db.Query(`SELECT shorts FROM URLs WHERE longs=$1`, longURL)
	if err != nil {
		return "", err
	}

	for rows.Next() {
		var shortURL string
		err = rows.Scan(&shortURL)
		return shortURL, err
	}
	return "", New("longURL")
}

func DBPut(db *sql.DB, longURL string, shortURL string) error {
	_, err := db.Exec(`insert into URLs(longs, shorts) values($1, $2)`, longURL, shortURL)
	return err
}

func GetMD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])[:7]
}

func DBCreateShortURL(db *sql.DB, longURL string) (string, error) {
	var shortURL string
	ll := longURL
	for {
		shortURL = GetMD5Hash(ll)
		fmt.Println(shortURL)
		_, err := DBGetLong(db, shortURL)
		if err != nil {
			if _, ok := err.(*DBNotFoundError); ok {
				break
			} else {
				return "", err
			}
		}
		ll += "0"
	}
	return shortURL, DBPut(db, longURL, shortURL)
}

type PutURL struct {
	LongURL string `json:"longurl" binding:"required"`
	// short string `json:"shorturl" binding:"required"`
}

func RecordGetTime(time float64) {
	getOpTime.Observe(time)
}

func RecordCreateTime(time float64) {
	createOpTime.Observe(time)
}

var (
	createOpTime = promauto.NewSummary(prometheus.SummaryOpts{
		Name: "create_op_time",
		Help: "Time of `create` operation processing",
	})

	getOpTime = promauto.NewSummary(prometheus.SummaryOpts{
		Name: "get_op_time",
		Help: "Time of `get` operation processing",
	})
)

// func recordMetrics() {
// 	go func() {
// 		for {
// 			opsProcessed.Inc()
// 			time.Sleep(2 * time.Second)
// 		}
// 	}()
// }

// var (
// 	opsProcessed = promauto.NewCounter(prometheus.CounterOpts{
// 		Name: "myapp_processed_ops_total",
// 		Help: "The total number of processed events",
// 	})
// )

func setupPrometheusRouter() *gin.Engine {
	r := gin.Default()
	registry := prometheus.NewRegistry()
	registry.MustRegister(createOpTime, getOpTime)
	handler := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})

	r.GET("/metrics", gin.WrapH(handler))
	return r
}

func setupRouter(db *sql.DB) *gin.Engine {
	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	// r.GET("/create/:url", func(c *gin.Context) {
	r.PUT("/create", func(c *gin.Context) {
		start := time.Now()

		requestBody := PutURL{}
		body, _ := io.ReadAll(c.Request.Body)
		err := json.Unmarshal(body, &requestBody)
		// err := c.BindJSON(&requestBody)
		if err != nil {
			c.String(http.StatusBadRequest, "")
			fmt.Println("Format error", err)
			return
		}

		fmt.Println(requestBody)
		longURL := requestBody.LongURL
		// longURL := c.Params.ByName("url")
		fmt.Println(longURL)
		result := http.StatusOK
		shortURL, err := DBGetShort(db, longURL)
		if err != nil {
			if _, ok := err.(*DBNotFoundError); ok {
				var err1 error
				shortURL, err1 = DBCreateShortURL(db, longURL)
				if err1 != nil {
					fmt.Println("db  error", err1)
					c.String(http.StatusBadRequest, "DB error")
				}
			} else {
				fmt.Println("db  error", err)
				c.String(http.StatusBadRequest, "DB error")
			}
		}

		c.JSON(result, gin.H{"longurl": longURL, "tinyurl": shortURL})

		elapsed := float64(time.Since(start).Nanoseconds()) / 1000

		RecordCreateTime(elapsed)		
	})

	r.GET("/:url", func(c *gin.Context) {
		start := time.Now()

		shortURL := c.Params.ByName("url")
		longURL, err := DBGetLong(db, shortURL)
		if err != nil {
			c.String(http.StatusNotFound, "")
			fmt.Println("There is no short url", err)
			return
		}
		c.Redirect(http.StatusFound, longURL)

		elapsed := float64(time.Since(start).Nanoseconds()) / 1000

		RecordGetTime(elapsed)
	})

	h := promhttp.Handler()

	r.GET("/metrics", func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
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
	defer conn.Close()

	err = conn.Ping()
	if err != nil {
		fmt.Println("Failed to ping database", err)
		panic("exit")
	}

	if _, err := conn.Exec(`create table if not exists urls (longs varchar, shorts varchar)`); err != nil {
		fmt.Println("Error creating table", err)
	}
	pr := setupPrometheusRouter()
	r := setupRouter(conn)

	go pr.Run(":8081")
	err = r.Run(":8080")
	if err != nil {
		panic(err.Error())
	}
}
