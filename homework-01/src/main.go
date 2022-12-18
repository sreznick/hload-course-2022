package main

import (
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"math/rand"
	"net/http"
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

// struct for user json request
type create struct {
	LONGURL string `json:"longurl"`
}

var links = []create{}

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

	http.Handle("/metrics", promhttp.Handler())
	go func() {
		http.ListenAndServe(":2112", nil)
	}()

	r := setupRouter()

	r.PUT("/create", func(c *gin.Context) { //User make put request
		var usersJSONLongUrl create //make new struct
		if err := c.BindJSON(&usersJSONLongUrl); err != nil {
			fmt.Println("Failed", err)
			panic("exit")
		}
		links = append(links, usersJSONLongUrl) //add to massive of slices
		linkBigFromUser := usersJSONLongUrl.LONGURL
		links = nil //delete last slice
		fmt.Println(links)
		var nameOfSearchingLinkInDB string
		err = conn.QueryRow("SELECT longurl FROM links WHERE longurl=$1", linkBigFromUser).Scan(&nameOfSearchingLinkInDB) //check is link in the db
		if err != nil {                                                                                                   //if db does not have a link
			rand.Seed(time.Now().UnixNano())
			charsAlphabetAndNumbers := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
				"abcdefghijklmnopqrstuvwxyz" +
				"0123456789")
			var buildTinyUrl strings.Builder
			for i := 0; i < 7; i++ {
				buildTinyUrl.WriteRune(charsAlphabetAndNumbers[rand.Intn(len(charsAlphabetAndNumbers))])
			} //build random tinyurl
			shortLinkForDB := buildTinyUrl.String()
			stmt, err := conn.Prepare("INSERT INTO links (longurl, tinyurl) VALUES ($1, $2);") //put links in db
			if err != nil {
				log.Fatal(err)
			}
			res, err := stmt.Exec(linkBigFromUser, shortLinkForDB)
			if err != nil {
				log.Fatal(err)
			}
			res = res
			defer stmt.Close()
			c.JSON(http.StatusOK, gin.H{"longurl": linkBigFromUser, "tinyurl": shortLinkForDB})
		} else { //if db has a link
			err = conn.QueryRow("SELECT tinyurl FROM links WHERE longurl=$1", linkBigFromUser).Scan(&nameOfSearchingLinkInDB)
			if err != nil {
				fmt.Println(err)
			}
			c.JSON(http.StatusOK, gin.H{"longurl": linkBigFromUser, "tinyurl": nameOfSearchingLinkInDB})
		}
	})

	r.GET("/:tiny", func(c *gin.Context) { //User make get request
		tiny := c.Params.ByName("tiny")
		var longurl string
		err = conn.QueryRow("SELECT longurl FROM links WHERE tinyurl =$1;", tiny).Scan(&longurl) //find longurl by tiny
		if err != nil {
			c.String(http.StatusNotFound, "404 page not found") //if link is out of db
		} else {
			c.Redirect(http.StatusFound, longurl) //redirect to page
			c.Abort()
		}
	})

	r.Run(":8080")
}
