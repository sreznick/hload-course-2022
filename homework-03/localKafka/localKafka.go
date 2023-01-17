package localKafka

import (
	"container/list"
	"context"
	"fmt"
	"localRedis"
	"strconv"

	"github.com/go-redis/redis/v8"
	"github.com/segmentio/kafka-go"
)

const (
	urlTopic       = "aisakova-tinyurls"
	clickTopic     = "aisakova-clicks"
	broker1Address = "158.160.19.212:9092"
	clickIncr      = 100
)

func CreateUrlWriter() *kafka.Writer {
	return kafka.NewWriter(kafka.WriterConfig{
		Brokers:      []string{broker1Address},
		Topic:        urlTopic,
		RequiredAcks: 1,
	})
}

func CreateClickWriter() *kafka.Writer {
	return kafka.NewWriter(kafka.WriterConfig{
		Brokers:      []string{broker1Address},
		Topic:        clickTopic,
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

func CreateClickReader() *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{broker1Address},
		Topic:   clickTopic,
	})
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

func ClickProduce(writer *kafka.Writer, ctx context.Context, clicks int) {
	err := writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(strconv.Itoa(clicks)),
		Value: []byte(strconv.Itoa(clicks)),
	})
	if err != nil {
		panic("could not write message " + err.Error())
	}
	fmt.Println("writes:", clicks)

}

func UrlConsume(reader *kafka.Reader, ctx context.Context, cluster *localRedis.RedisCluster, id int) {
	(*cluster).RedisOptions.Addr = (*cluster).Workers[id]
	rdb := redis.NewClient(&(*cluster).RedisOptions)
	for {
		msg, err := reader.ReadMessage(ctx)
		if err != nil {
			panic("could not read message " + err.Error())
		}
		rdb.Do(ctx, "set", (*cluster).Prefix+"_"+string(msg.Key), string(msg.Value))
		fmt.Println("received: ", string(msg.Key))

	}
}

func ClickConsume(reader *kafka.Reader, ctx context.Context, cluster *localRedis.RedisCluster, tinyUrl string) {
	(*cluster).RedisOptions.Addr = (*cluster).Master
	rdb := redis.NewClient(&(*cluster).RedisOptions)
	for {
		msg, err := reader.ReadMessage(ctx)
		if err != nil {
			panic("could not read message " + err.Error())
		}
		rdb.Do(ctx, "incrby", (*cluster).Prefix+"_"+tinyUrl+"_"+"clicks", clickIncr)

		fmt.Println("received: ", string(msg.Value))

	}

}
