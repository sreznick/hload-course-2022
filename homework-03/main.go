package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/segmentio/kafka-go"
	"log"
	"math/rand"
	"net/http"
	"strings"
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

// struct for user json request
type create struct {
	LONGURL string `json:"longurl"`
}

var links = []create{}

func main() {

	writer := &kafka.Writer{
		Addr:  kafka.TCP("158.160.19.212:9092"),
		Topic: "mdiagilev-test",
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   []string{"158.160.19.212:9092"},
		Topic:     "mdiagilev-test",
		Partition: 0,
	})

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

			//produce
			error := writer.WriteMessages(context.Background(),
				kafka.Message{
					Key:   []byte(linkBigFromUser),
					Value: []byte(shortLinkForDB),
				},
			)

			if error != nil {
				log.Fatal("failed to write messages:", error)
			}

			if error := writer.Close(); error != nil {
				log.Fatal("failed to close writer:", error)
			}

			//consume
			go func() {
				m, errorerrr := reader.ReadMessage(context.Background())
				if errorerrr != nil {
					fmt.Println("1")
				}
				fmt.Printf("message: %s = %s\n", string(m.Key), string(m.Value))
			}()

			if errorerrr := reader.Close(); errorerrr != nil {
				log.Fatal("failed to close reader:", errorerrr)
			}

		} else { //if db has a link
			err = conn.QueryRow("SELECT tinyurl FROM links WHERE longurl=$1", linkBigFromUser).Scan(&nameOfSearchingLinkInDB)
			if err != nil {
				fmt.Println(err)
			}

			// produce
			error := writer.WriteMessages(context.Background(),
				kafka.Message{
					Key:   []byte(linkBigFromUser),
					Value: []byte(nameOfSearchingLinkInDB),
				},
			)

			if error != nil {
				log.Fatal("failed to write messages:", error)
			}

			if error := writer.Close(); error != nil {
				log.Fatal("failed to close writer:", error)
			}

			//consume
			go func() {
				m, errorerr := reader.ReadMessage(context.Background())
				if errorerr != nil {
					fmt.Println("2")
				}
				fmt.Printf("message: %s = %s\n", string(m.Key), string(m.Value))
			}()

			if errorerr := reader.Close(); errorerr != nil {
				log.Fatal("failed to close reader:", errorerr)
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
			fmt.Println(longurl)
			//c.Redirect(http.StatusFound, longurl) //redirect to page
			//c.Abort()
		}
	})

	r.Run("0.0.0.0:26379")

}
