package localKafka

import (
	"container/list"
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
	"github.com/segmentio/kafka-go"
)

const (
	urlTopic       = "aisakova-test"
	broker1Address = "158.160.19.212:9092"
)

func CreateUrlWriter() *kafka.Writer {
	return kafka.NewWriter(kafka.WriterConfig{
		Brokers:      []string{broker1Address},
		Topic:        urlTopic,
		RequiredAcks: 1,
	})
}

func CreateUrlReaders(nWorkers int) *list.List {
	urlReaders := list.New()
	for i := 0; i < nWorkers; i++ {
		urlReaders.PushBack(kafka.NewReader(kafka.ReaderConfig{
			Brokers: []string{broker1Address},
			Topic:   urlTopic,
		}))

	}
	return urlReaders

}

func UrlProduce(writer *kafka.Writer, ctx context.Context, longUrl string, tinyUrl string) {
	err := writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(tinyUrl),
		Value: []byte(longUrl),
	})
	if err != nil {
		panic("could not write message " + err.Error())
	}
	fmt.Println("writes:", tinyUrl+":"+longUrl)

}

func UrlConsume(reader *kafka.Reader, ctx context.Context) {
	redisOptions := redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	}
	rdb := redis.NewClient(&redisOptions)
	for {
		msg, err := reader.ReadMessage(ctx)
		if err != nil {
			panic("could not read message " + err.Error())
		}
		rdb.Do(ctx, "set", string(msg.Key), string(msg.Value))
		fmt.Println("received: ", string(msg.Value))

	}
}
