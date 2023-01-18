package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/segmentio/kafka-go"
	"io/ioutil"
	"log"
	"main/worker"
	"math/rand"
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

func ReplicaReadNewDataFromMaster(topic string) {
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
		//записываем
		fmt.Printf("message with links from Master to replica: %s = %s\n", string(m.Key), string(m.Value))
		for i := 0; i < len(getHost()); i++ {

			redis := worker.Redis{Cluster: getHost()[i]}
			err = redis.Connect()
			if err != nil {
				panic(err)
			}
			defer redis.Close()
			ctx := context.Background()
			redis.Connect()
			error := redis.Client.HSet(ctx, "mdyagilev:main", string(m.Key), string(m.Value))
			if error != nil {
				fmt.Println("value not record")
			}
		}
		if err := reader.Close(); err != nil {
			log.Fatal("failed to close reader:", err)
		}
	}

}

func getHost() []string {
	data, err := ioutil.ReadFile("ssh_hosts")
	if err != nil {
		fmt.Println(err)
	}
	sshHosts := string(data)
	return strings.Split(sshHosts, " ")

}

func main() {
	r := setupRouter()

	r.GET("/:tiny", func(c *gin.Context) { //User make get request
		tiny := c.Params.ByName("tiny")

		rand.Seed(time.Now().UnixNano())
		var randomNumber = rand.Int63n(2)

		redis := worker.Redis{Cluster: getHost()[randomNumber]}
		err := redis.Connect()
		if err != nil {
			panic(err)
		}
		defer redis.Close()

		ctx := context.Background()
		valueOfLongUrl, err := redis.Client.HGetAll(ctx, "mdyagilev:main").Result()
		if err != nil {
			fmt.Println("", err)
		}

		val, ok := valueOfLongUrl[tiny]
		if ok == false {
			c.Writer.WriteHeader(http.StatusNotFound)
		} else {
			c.Redirect(http.StatusFound, val) // если ссылка есть в redis
		}
	})

	go ReplicaReadNewDataFromMaster("mdiagilev-test-master")

	r.Run("0.0.0.0:8081")
}
