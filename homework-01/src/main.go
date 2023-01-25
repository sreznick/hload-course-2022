package main

import (
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	_ "strings"
	"time"
)

const SQL_DRIVER = "postgres"
const SQL_CONNECT_URL = "postgres://postgres:postgrespw@localhost:49153/postgres?sslmode=disable"

func setupRouter() *gin.Engine {
	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	return r
}

func generateTinyUrl(conn *sql.DB) string {
	for {
		rand.Seed(time.Now().UnixNano())
		charsAlphabetAndNumbers := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
			"abcdefghijklmnopqrstuvwxyz" +
			"0123456789")
		var buildTinyUrl strings.Builder
		for i := 0; i < 7; i++ {
			buildTinyUrl.WriteRune(charsAlphabetAndNumbers[rand.Intn(len(charsAlphabetAndNumbers))])
		} //build random tinyurl

		url := conn.QueryRow("SELECT longurl FROM links WHERE tinyurl=$1", buildTinyUrl.String()) // check collides
		if url != nil {
			return buildTinyUrl.String()
		}

	}
}

func getHosts() string {
	data, err := os.ReadFile("port")
	if err != nil {
		fmt.Print(err)
	}

	return string(data)
}

// struct for user json request
type create struct {
	LongUrl string `json:"longurl"`
}

var MetricOnGET = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name: "get_requests",
		Help: "Количество запросов на переход по shortUrl",
	})

var MetricOnPUT = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name: "put_requests",
		Help: "Количество запросов на создание shortUrl по longUrl",
	})

func main() {
	conn, err := sql.Open(SQL_DRIVER, SQL_CONNECT_URL)
	conn.SetMaxIdleConns(10000)
	if err != nil {
		fmt.Println("Failed to open", err)
		panic("exit")
	}

	err = conn.Ping()
	if err != nil {
		fmt.Println("Failed to ping database", err)
		panic("exit")
	}

	prometheus.MustRegister(MetricOnGET)
	prometheus.MustRegister(MetricOnPUT)
	http.Handle("/metrics", promhttp.Handler())
	go func() {
		http.ListenAndServe(":2112", nil)

	}()

	r := setupRouter()

	r.PUT("/create", func(ctx *gin.Context) { //User make put request
		MetricOnPUT.Inc()
		var (
			usersJSONLongUrl create //make new struct
		)
		if err := ctx.BindJSON(&usersJSONLongUrl); err != nil {
			fmt.Println("Failed", err)
			panic("exit")
		}

		linkBigFromUser := usersJSONLongUrl.LongUrl
		var nameOfSearchingLinkInDB string
		err = conn.QueryRow("SELECT longurl FROM links WHERE longurl=$1", linkBigFromUser).Scan(&nameOfSearchingLinkInDB) //check is link in the db
		if err != nil {
			//if db does not have a link
			shortLinkForDB := generateTinyUrl(conn) // generate tinyurl
			// start transaction
			tx, err := conn.BeginTx(ctx, nil)
			if err != nil {
				log.Fatal(err)
			}

			stmt, err := tx.Prepare("INSERT INTO links (longurl, tinyurl) VALUES ($1, $2);") //put links in db
			if err != nil {
				tx.Rollback()
				return
			}
			defer stmt.Close()
			_, err = stmt.Exec(linkBigFromUser, shortLinkForDB)
			if err != nil {
				log.Fatal(err)
			}
			err = tx.Commit() // end transaction
			if err != nil {
				log.Fatal(err)
			}

			ctx.JSON(http.StatusOK, gin.H{"longurl": linkBigFromUser, "tinyurl": shortLinkForDB})
		} else {
			//if db has a link
			err = conn.QueryRow("SELECT tinyurl FROM links WHERE longurl=$1", linkBigFromUser).Scan(&nameOfSearchingLinkInDB)
			if err != nil {
				fmt.Println(err)
			}
			ctx.JSON(http.StatusOK, gin.H{"longurl": linkBigFromUser, "tinyurl": nameOfSearchingLinkInDB})
		}

	})

	r.GET("/:tiny", func(ctx *gin.Context) { //User make get request
		MetricOnGET.Inc()
		tiny := ctx.Params.ByName("tiny")
		var longurl string
		err = conn.QueryRow("SELECT longurl FROM links WHERE tinyurl =$1;", tiny).Scan(&longurl) //find longurl by tiny
		if err != nil {
			ctx.String(http.StatusNotFound, "404 page not found") //if link is out of db
		} else {
			ctx.Redirect(http.StatusFound, longurl) //redirect to page
			ctx.Abort()
		}

	})

	r.Run(getHosts())
}
