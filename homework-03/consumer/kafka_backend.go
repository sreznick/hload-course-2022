package consumer

import (
	"context"
	"fmt"
	"main/common"
	"main/consumer/kafka"
)

func unpackUrlMsg(b []byte) (string, string) {
	p := string(b)
	d := 0
	for i := 0; i < len(p); i++ {
		if p[i] == common.Delim {
			d = i
			break
		}
	}

	return p[0:d], p[d+1:]
}

func KafkaDvij() {
	ctx := context.Background()
	for {
		v, err := kafka.UrlsConsumer.FetchMessage(ctx)

		if err != nil {
			fmt.Printf("Kafka error: %s\n", err)
			continue
		}

		t, l := unpackUrlMsg(v.Value)
		err = SetLongUrl(t, l)
		if err != nil {
			fmt.Printf("Something went wrong with Redis: %s\n", err)
			continue
		}

		err = CreateClick(t)
		if err != nil {
			fmt.Printf("Something went wrong with Redis: %s\n", err)
			continue
		}

		err = IncrementClick(t)
		if err != nil {
			fmt.Printf("Something went wrong with Redis: %s\n", err)
			continue
		}

		err = kafka.UrlsConsumer.CommitMessages(ctx, v)
		if err != nil {
			fmt.Printf("Something went wrong with Redis: %s\n", err)
			continue
		}

		fmt.Printf("New url recieved -- %s, %s\n", t, l)
	}
}
