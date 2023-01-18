package main

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"github.com/gin-gonic/gin"
	_ "github.com/go-redis/redis/v8"
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

func MasterSendNewDataToReplicas(longLink, shortLink, topicToReplica string) {
	writer := &kafka.Writer{
		Addr:  kafka.TCP("158.160.19.212:9092"),
		Topic: topicToReplica}

	error1 := writer.WriteMessages(context.Background(),
		kafka.Message{
			Key:   []byte(shortLink), //longurl from db
			Value: []byte(longLink),  //tinyurl from replica
		})
	if error1 != nil {
		log.Fatal("failed to write messages:1:", error1)
	}
	if error1 := writer.Close(); error1 != nil {
		log.Fatal("failed to close writer:", error1)
	}

}

func MakeTinyUrl(conn *sqlx.DB, linkBigFromUser string, err error) (result bool, resultLink string) {
	var nameOfSearchingLinkInDB string
	err = conn.QueryRow("SELECT longurl FROM links WHERE longurl=$1", linkBigFromUser).Scan(&nameOfSearchingLinkInDB) //check is link in the db
	if err == nil {
		//if db has a link
		result = false
		err = conn.QueryRow("SELECT tinyurl FROM links WHERE longurl=$1", linkBigFromUser).Scan(&nameOfSearchingLinkInDB)
		if err != nil {
			fmt.Println(err)
		}
		resultLink = nameOfSearchingLinkInDB
		return
	} //if db does not have a link

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
	_, err = stmt.Exec(linkBigFromUser, shortLinkForDB)
	if err != nil {
		log.Fatal(err)
	}

	defer stmt.Close()
	return
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

		if result == true { //if db does not have a link
			MasterSendNewDataToReplicas(linkBigFromUser, resultLink, "mdiagilev-test-master")
			//go ReplicaReadNewDataFromMaster("mdiagilev-test-master")
			c.JSON(http.StatusOK, gin.H{"longurl": linkBigFromUser, "tinyurl": resultLink})
		} else { //if db has a link
			c.JSON(http.StatusOK, gin.H{"longurl": linkBigFromUser, "tinyurl": resultLink})
		}
	})
	r.Run("0.0.0.0:8080")
}
