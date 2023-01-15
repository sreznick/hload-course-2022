package consumer

import (
	"context"
	"fmt"
	"github.com/segmentio/kafka-go"
	kafka2 "main/consumer/kafka"
)

func PushClicks(tinyUrl string) {
	// Produce messages to topic (asynchronously)
	ctx := context.Background()
	err := kafka2.ClicksProducer.WriteMessages(ctx, kafka.Message{
		Key: []byte(tinyUrl),
	})
	if err != nil {
		fmt.Printf("Error while clicks produce: %s\n", err)
		return
	}

	fmt.Printf("Successfully produced clicks")
}
