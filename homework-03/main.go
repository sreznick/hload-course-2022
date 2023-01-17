package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/segmentio/kafka-go"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

const SQL_DRIVER = "postgres"

//const SQL_CONNECT_URL = "postgres://postgres:postgrespw@localhost:49153/postgres?sslmode=disable"

const SQL_CONNECT_URL = "postgres://postgres:@localhost:22/postgres?sslmode=disable"

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

	//writer := &kafka.Writer{
	//	Addr:  kafka.TCP("158.160.19.212:9092"),
	//	Topic: "mdiagilev-test"}

	//reader := kafka.NewReader(kafka.ReaderConfig{
	//	Brokers:   []string{"158.160.19.212:9092"},
	//	Topic:     "mdiagilev-test",
	//	Partition: 0})
	//reader.SetOffset(kafka.LastOffset)

	fmt.Println(sql.Drivers())
	//b, err := ioutil.ReadFile("pass.conf")
	//if err != nil {
	//	panic(err)
	//}

	key, err := ioutil.ReadFile("//users//bogdan//.ssh//id_ed25519")
	if err != nil {
		panic(err)
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		panic(err)
	}

	// convert bytes to string
	pass := "postgres"

	server := &SSH{
		Ip:     "51.250.106.140",
		User:   "mdiagilev",
		Port:   22,
		Cert:   pass,
		Signer: signer,
	}

	err = server.Connect(CERT_PUBLIC_KEY_FILE)
	if err != nil {
		panic(err)
	}

	defer server.Close()
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
			//go func() {
			writer := &kafka.Writer{
				Addr:  kafka.TCP("158.160.19.212:9092"),
				Topic: "mdiagilev-test"}

			error1 := writer.WriteMessages(context.Background(),
				kafka.Message{
					Key:   []byte(linkBigFromUser),
					Value: []byte(shortLinkForDB),
				})
			if error1 != nil {
				log.Fatal("failed to write messages:1:", error1)
			}
			if error1 := writer.Close(); error1 != nil {
				log.Fatal("failed to close writer:", error1)
			}

			//consume
			go func() {
				reader := kafka.NewReader(kafka.ReaderConfig{
					Brokers:   []string{"158.160.19.212:9092"},
					Topic:     "mdiagilev-test",
					Partition: 0})
				reader.SetOffset(kafka.LastOffset)

				m, errorerrr := reader.ReadMessage(context.Background())
				if errorerrr != nil {
					fmt.Println("1")
				}
				fmt.Printf("message: %s = %s\n", string(m.Key), string(m.Value))
				if errorerrr := reader.Close(); errorerrr != nil {
					log.Fatal("failed to close reader:", errorerrr)
				}
			}()
		} else { //if db has a link
			err = conn.QueryRow("SELECT tinyurl FROM links WHERE longurl=$1", linkBigFromUser).Scan(&nameOfSearchingLinkInDB)
			if err != nil {
				fmt.Println(err)
			}

			// produce
			writer := &kafka.Writer{
				Addr:  kafka.TCP("158.160.19.212:9092"),
				Topic: "mdiagilev-test"}
			error2 := writer.WriteMessages(context.Background(),
				kafka.Message{
					Key:   []byte(linkBigFromUser),
					Value: []byte(nameOfSearchingLinkInDB),
				},
			)

			if error2 != nil {
				log.Fatal("failed to write messages:", error2)
			}

			if error2 := writer.Close(); error2 != nil {
				log.Fatal("failed to close writer:", error2)
			}

			//consume
			go func() {
				reader := kafka.NewReader(kafka.ReaderConfig{
					Brokers:   []string{"158.160.19.212:9092"},
					Topic:     "mdiagilev-test",
					Partition: 0})
				reader.SetOffset(kafka.LastOffset)

				m, errorerr := reader.ReadMessage(context.Background())
				if errorerr != nil {
					fmt.Println("2")
				}
				fmt.Printf("message: %s = %s\n", string(m.Key), string(m.Value))
				if errorerr := reader.Close(); errorerr != nil {
					log.Fatal("failed to close reader:", errorerr)
				}
			}()

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

	r.Run("0.0.0.0:8080")

}
