package worker

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/segmentio/kafka-go"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
)

func randomNumber() int64 {
	rand.Seed(time.Now().UnixNano())
	return rand.Int63n(2)
}

func Get(ctx *gin.Context) error {
	var (
		click int64
		tiny  = ctx.Params.ByName("tiny")
	)
	hosts, err := getHosts()
	if err != nil {
		return fmt.Errorf("get.getHosts: %w", err)
	}

	redis := Redis{
		Cluster: hosts[randomNumber()],
	}
	if err = redis.Connect(); err != nil {
		log.Fatalf("redis.Connect: %v", err)
	}
	defer redis.Close()

	valueOfLongUrl, err := redis.HGetAll(ctx, "mdyagilev:main").Result()
	if err != nil {
		return fmt.Errorf("redis.HGetAll: %w", err)
	}

	val, ok := valueOfLongUrl[tiny]
	if !ok {
		ctx.Writer.WriteHeader(http.StatusNotFound) // если в redis реплике нет ссылка
		return nil
	}

	ctx.Redirect(http.StatusFound, val)                                  // если в redis реплике есть ссылка
	click, err = redis.HIncrBy(ctx, "mdyagilev:clicks", val, 1).Result() //  Incr(ctx, val).Result()
	if err != nil {
		return fmt.Errorf("redis.HIncrBy: %w", err)
	}

	if click == 10 { //100
		if err = replicaSendToMasterClickOnURL(
			ctx,
			val,
			"mdiagilev-test",
			fmt.Sprintf("%v", click),
		); err != nil {
			return err
		}
	}

	return nil
}

func SetupRouter() *gin.Engine {
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	return r
}

func ReplicaReadNewDataFromMaster(ctx context.Context, topic string) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   []string{"158.160.19.212:9092"},
		Topic:     topic,
		Partition: 0})
	err := reader.SetOffset(kafka.LastOffset)
	if err != nil {
		log.Fatalf("reader.SetOffset: %v", err)
	}
	defer reader.Close()

	for {
		select {
		case <-ctx.Done():
			log.Printf("contex done")
		default:
			if err = readNewDataFromMaster(ctx, reader); err != nil {
				log.Fatalf("readNewDataFromMaster: %v", err)
			}
		}
	}
}

func readNewDataFromMaster(ctx context.Context, reader *kafka.Reader) error {
	message, err := reader.ReadMessage(ctx)
	if err != nil {
		return fmt.Errorf("reader.ReadMessage: %w", err)
	}

	var hosts []string
	hosts, err = getHosts()
	for _, host := range hosts {
		redis := Redis{Cluster: host}
		if err = redis.Connect(); err != nil {
			return err
		}
		defer redis.Close()

		init := redis.HSet(ctx, "mdyagilev:main", string(message.Key), string(message.Value))
		if init != nil {
			log.Printf("value not record")
		}

		init = redis.HSet(ctx, "mdyagilev:clicks", string(message.Value), 0)
		if init != nil {
			log.Printf("value not record")
		}

	}

	return nil
}

func replicaSendToMasterClickOnURL(ctx context.Context, longLink, topicToReplica, clickNumber string) error {
	writer := &kafka.Writer{
		Addr:  kafka.TCP("158.160.19.212:9092"),
		Topic: topicToReplica,
	}

	err := writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(longLink),    //longurl from db
		Value: []byte(clickNumber), //tinyurl from replica
	})
	if err != nil {
		return fmt.Errorf("failed to write messages: %w", err)
	}

	if err = writer.Close(); err != nil {
		return fmt.Errorf("failed to close writer: %w", err)
	}

	return nil
}

func getHosts() ([]string, error) {
	data, err := os.ReadFile("ssh_hosts")
	if err != nil {
		return nil, err
	}

	return strings.Split(string(data), " "), nil
}
