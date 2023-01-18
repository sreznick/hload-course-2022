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
	main2 "main/master/src"
	"main/worker/src"
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

var sshcon main2.SSH

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
		reader.SetOffset(kafka.FirstOffset)
		m, errorerr := reader.ReadMessage(context.Background())
		if errorerr != nil {
			fmt.Println("2")
		}
		//записываем в рэдис, переменную под рэдис сделать

		fmt.Printf("message: %s = %s\n", string(m.Key), string(m.Value))
		if errorerr := reader.Close(); errorerr != nil {
			log.Fatal("failed to close reader:", errorerr)
		}

	}
}

func MasterReadFromTopic(conn *sqlx.DB, err error, topic string) (string, string) {
	var link string
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   []string{"158.160.19.212:9092"},
		Topic:     topic,
		Partition: 0})
	reader.SetOffset(kafka.LastOffset)
	// пишем мастеру longurl которой нет в redis

	m, errorerr := reader.ReadMessage(context.Background())
	if errorerr != nil {
		fmt.Println("2")
	}
	fmt.Printf("message from replica to Master: %s = %s\n", string(m.Key), string(m.Value))
	if errorerr := reader.Close(); errorerr != nil {
		log.Fatal("failed to close reader:", errorerr)
	}
	var nameOfSearchingLinkInDB string
	nameOfSearchingLinkInDB = string(m.Key)
	err = conn.QueryRow("SELECT longurl FROM links WHERE tinyurl=$1", string(m.Key)).Scan(&nameOfSearchingLinkInDB)
	if err != nil {
		fmt.Println(err)
	}

	link = nameOfSearchingLinkInDB
	return link, string(m.Key) // большая, small
}

func MasterSendToTopic(longLink, shortLink, topicToReplica string) {
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

func ReplicaReadResponseOnHerRequest(redis src.Redis, ctx context.Context, topic string) {

	for {
		reader := kafka.NewReader(kafka.ReaderConfig{
			Brokers:   []string{"158.160.19.212:9092"},
			Topic:     topic,
			Partition: 0})
		reader.SetOffset(kafka.LastOffset)
		m, err := reader.ReadMessage(context.Background())
		if err != nil {
			fmt.Println("2")
		}
		//записываем в рэдис, переменную под рэдис сделать
		fmt.Printf("message with links from Master to replica: %s = %s\n", string(m.Key), string(m.Value))

		redis.Connect()
		error := redis.Client.HSet(ctx, "mdyagilev:main", string(m.Key), string(m.Value))
		if error != nil {
			fmt.Println("value not record")
		}

		if err := reader.Close(); err != nil {
			log.Fatal("failed to close reader:", err)
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

func getHost() string {
	data, err := ioutil.ReadFile("ssh_hosts")
	if err != nil {
		fmt.Println(err)
	}
	dataString := string(data)
	value := strings.Split(dataString, " ")
	rand.Seed(time.Now().UnixNano())
	var randomNumber = rand.Int63n(2)
	if randomNumber == 0 {
		return value[0]
	} else {
		return value[0] //1!!!!!!!
	}
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

	server := &main2.SSH{
		Ip:     "51.250.106.140",
		User:   "mdiagilev",
		Port:   22,
		Cert:   pass,
		Signer: signer,
	}

	err = server.Connect(main2.CERT_PUBLIC_KEY_FILE)
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

		//check is link in the db
		if result == true { //if db does not have a link
			//go Writer(linkBigFromUser, resultLink, "mdiagilev-test") //5.send url
			c.JSON(http.StatusOK, gin.H{"longurl": linkBigFromUser, "tinyurl": resultLink})
		} else { //if db has a link

			c.JSON(http.StatusOK, gin.H{"longurl": linkBigFromUser, "tinyurl": resultLink})
		}
		//	redis
	})

	//go Writer(linkBigFromUser, resultLink, "mdiagilev-test")
	//go Read("mdiagilev-test-master")

	//go ReplicaReadResponseOnHerRequest("mdiagilev-test")

	var redis src.Redis
	var ctx context.Context
	go ReplicaReadResponseOnHerRequest(redis, ctx, "mdiagilev-test")

	r.GET("/:tiny", func(c *gin.Context) { //User make get request
		tiny := c.Params.ByName("tiny")

		redis := src.Redis{Cluster: getHost()}
		err = redis.Connect()
		if err != nil {
			panic(err)
		}
		defer redis.Close()

		//прочитать все что есть в топике если мы умерли,
		ctx := context.Background()
		valueOfLongUrl, err := redis.Client.HGetAll(ctx, "mdyagilev:main").Result()
		if err != nil {
			fmt.Println("", err)
		}

		val, ok := valueOfLongUrl[tiny]
		if ok == false {
			c.Writer.WriteHeader(http.StatusNotFound)
			//если ссылки нет в redis
			WriterToMaster(tiny, "mdiagilev-test-master")                                   //если в нем нет то пишем в топик мастера
			var longUrl, shortUrl = MasterReadFromTopic(conn, err, "mdiagilev-test-master") // мастер прочитает с
			// топика и пойдет в бд, заберет значение
			MasterSendToTopic(longUrl, shortUrl, "mdiagilev-test") //затем отправит в топик мастер это значение
			//go ReplicaReadResponseOnHerRequest(redis, ctx, "mdiagilev-test") //читает сообщения от мастера
		} else {
			c.Redirect(http.StatusFound, val) // если ссылка есть в redis
		}

		//____________________________________________________________________________________________________//
		//если в нем нет то пишем в топик мастера
		//go WriterToMaster(tiny, "mdiagilev-test-master") // 3. request url
		////мастер прочитает с топика мастера и пойдет в бд, заберет и вернет значение
		//var longurl string
		//var shorturl string
		//longurl, shorturl = MasterReadFromTopic(conn, err, "mdiagilev-test-master")
		////затем отправит в топик мастер это значение
		//MasterSendToTopic(longurl, shorturl, "mdiagilev-test")
		//реплика прочитает топик и заберет последнее
		//go ReplicaReadResponseOnHerRequest("mdiagilev-test")
		//читает из топика дальше и пишет в рэдис
		//____________________________________________________________________________________________________//
		//идем и делаем запрос в рэдис

		//go Read("mdiagilev-test-master")
	})

	//go ReplicaReadResponseOnHerRequest(redis, ctx, "mdiagilev-test")

	r.Run("0.0.0.0:8080")

}
