package consumer

import (
	"context"
	"fmt"
	"github.com/segmentio/kafka-go"
)

func PushClicks(tinyUrl string) {
	// Produce messages to topic (asynchronously)
	ctx := context.Background()
	err := ClicksProducer.WriteMessages(ctx, kafka.Message{
		Key: []byte(tinyUrl),
	})
	if err != nil {
		fmt.Printf("Error while clicks produce: %s\n", err)
		return
	}

	fmt.Printf("Successfully produced clicks")
}
