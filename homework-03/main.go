package main

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
	"github.com/segmentio/kafka-go"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"net/http"
	"strings"
	"time"
)

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

var sshcon SSH

type ViaSSHDialer struct {
	client *ssh.Client
}

func (self *ViaSSHDialer) Open(s string) (_ driver.Conn, err error) {
	return pq.DialOpen(self, s)
}

func (self *ViaSSHDialer) Dial(network, address string) (net.Conn, error) {
	return self.client.Dial(network, address)
}

func (self *ViaSSHDialer) DialTimeout(network, address string, timeout time.Duration) (net.Conn, error) {
	return self.client.Dial(network, address)
}

func Read(topic string) {
	var res = true
	for res {
		reader := kafka.NewReader(kafka.ReaderConfig{
			Brokers:   []string{"158.160.19.212:9092"},
			Topic:     topic,
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
	}
}

func Writer(linkBigFromUser, shortLinkForDB, topic string) {
	writer := &kafka.Writer{
		Addr:  kafka.TCP("158.160.19.212:9092"),
		Topic: topic}

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
}

func WriterToMaster(shortLinkForDB, topic string) {
	writer := &kafka.Writer{
		Addr:  kafka.TCP("158.160.19.212:9092"),
		Topic: topic}

	error1 := writer.WriteMessages(context.Background(),
		kafka.Message{
			Key:   []byte(shortLinkForDB),
			Value: []byte(shortLinkForDB),
		})
	if error1 != nil {
		log.Fatal("failed to write messages:1:", error1)
	}
	if error1 := writer.Close(); error1 != nil {
		log.Fatal("failed to close writer:", error1)
	}
}

func MakeTinyUrl(conn *sqlx.DB, linkBigFromUser string, err error) (bool, string) {
	var nameOfSearchingLinkInDB string
	var result bool
	var resultLink string
	err = conn.QueryRow("SELECT longurl FROM links WHERE longurl=$1", linkBigFromUser).Scan(&nameOfSearchingLinkInDB) //check is link in the db
	if err != nil {                                                                                                   //if db does not have a link
		result = true
		rand.Seed(time.Now().UnixNano())
		charsAlphabetAndNumbers := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
			"abcdefghijklmnopqrstuvwxyz" +
			"0123456789")
		var buildTinyUrl strings.Builder
		for i := 0; i < 7; i++ {
			buildTinyUrl.WriteRune(charsAlphabetAndNumbers[rand.Intn(len(charsAlphabetAndNumbers))])
		} //build random tinyurl
		shortLinkForDB := buildTinyUrl.String()
		resultLink = shortLinkForDB
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

	} else { //if db has a link
		result = false
		err = conn.QueryRow("SELECT tinyurl FROM links WHERE longurl=$1", linkBigFromUser).Scan(&nameOfSearchingLinkInDB)
		if err != nil {
			fmt.Println(err)
		}
		resultLink = nameOfSearchingLinkInDB
	}

	return result, resultLink
}

func main() {

	key, err := ioutil.ReadFile("//users//sergeidiagilev//.ssh//id_ed25519")
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

	sql.Register("postgres+ssh", &ViaSSHDialer{server.client})

	conn, err := sqlx.Open("postgres+ssh", "user=postgres dbname=mdiagilev sslmode=disable")
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

	go Read("mdiagilev-test")
	//consumer master
	go Read("mdiagilev-test-master")

	r.PUT("/create", func(c *gin.Context) { //User make put request
		var usersJSONLongUrl create //make new struct
		if err := c.BindJSON(&usersJSONLongUrl); err != nil {
			fmt.Println("Failed", err)
			panic("exit")
		}
		linkBigFromUser := usersJSONLongUrl.LONGURL
		var result bool
		var resultLink string
		result, resultLink = MakeTinyUrl(conn, linkBigFromUser, err)

		//check is link in the db
		if result == true { //if db does not have a link
			go Writer(linkBigFromUser, resultLink, "mdiagilev-test")
			c.JSON(http.StatusOK, gin.H{"longurl": linkBigFromUser, "tinyurl": resultLink})
		} else { //if db has a link

			c.JSON(http.StatusOK, gin.H{"longurl": linkBigFromUser, "tinyurl": resultLink})
		}
		//	redis
	})

	//go Writer(linkBigFromUser, resultLink, "mdiagilev-test")

	r.GET("/:tiny", func(c *gin.Context) { //User make get request
		//redis
		tiny := c.Params.ByName("tiny")
		go WriterToMaster(tiny, "mdiagilev-test-master")
	})

	r.Run("0.0.0.0:8080")

}
