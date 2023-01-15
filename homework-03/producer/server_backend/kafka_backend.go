package server_backend

import (
	"context"
	"database/sql"
	"fmt"
	"main/common"
	"main/producer/url_backend"
	"strconv"

	"github.com/segmentio/kafka-go"
)

var urlsProducer kafka.Writer

var clicksReader kafka.Reader

func getUpdateQuery() string {
	return "update urls set clicks = clicks + " + strconv.Itoa(common.ClicksThrsh) + " WHERE ID = $1;"
}

func SetProducerKafka(c common.KafkaConfig) {
	clicksReader = *kafka.NewReader(kafka.ReaderConfig{
		Brokers:   c.ClicksBrokers,
		GroupID:   c.ClicksGroup,
		Topic:     c.ClicksTopicName,
		Partition: 0,
	})

	urlsProducer = kafka.Writer{
		Addr:     kafka.TCP(c.UrlsProducing),
		Topic:    c.UrlTopicName,
		Balancer: &kafka.LeastBytes{},
	}
}

func packUrlMsg(tinyUrl string, longUrl string) []byte {
	return []byte(tinyUrl + string(common.Delim) + longUrl)
}

func PushCreatedUrl(tinyUrl string, longUrl string) {
	// Produce messages to topic (asynchronously)
	ctx := context.Background()
	err := urlsProducer.WriteMessages(ctx, kafka.Message{
		Value: packUrlMsg(tinyUrl, longUrl),
	})
	if err != nil {
		fmt.Printf("Error while clicks produce: %s", err)
		return
	}

	fmt.Printf("Successfully produced url")
}

// Read kafka url topic
func GetClicks(db *sql.DB) {
	ctx := context.Background()
	for {
		v, err := clicksReader.FetchMessage(ctx)

		if err != nil {
			fmt.Printf("Kafka error: %s\n", err)
			continue
		}

		t, err := url_backend.UrlToId(string(v.Key))
		if err != nil {
			fmt.Printf("Something went wrong with Clicks parsing: %s\n", err)
			continue
		}

		_, err = db.Exec(getUpdateQuery(), t)
		if err != nil {
			fmt.Printf("Something went wrong with Clicks in DB: %s\n", err)
			continue
		}

		err = clicksReader.CommitMessages(ctx, v)
		if err != nil {
			fmt.Printf("Kafka error: %s\n", err)
		}

		fmt.Printf("New clicks recieved -- %s, %d\n", v.Key, common.ClicksThrsh)
	}
}
