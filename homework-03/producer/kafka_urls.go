package producer

import (
	"context"
	"fmt"
	"github.com/segmentio/kafka-go"
	"main/common"
)

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
