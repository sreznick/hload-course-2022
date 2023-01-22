package countclicks

import (
	"context"
	"encoding/json"
	"log"
	"main/internal/config"
	"main/internal/kafka"
	"main/internal/models"
	"main/internal/postgres"
	"time"

	"github.com/Shopify/sarama"
)

type countClicks struct {
	client   sarama.ConsumerGroup
	consumer kafka.Consumer
	postgres postgres.Interface
}

func NewCountClicksJob(postgres postgres.Interface) *countClicks {
	saramaConfig := sarama.NewConfig()
	saramaConfig.Consumer.Offsets.Initial = sarama.OffsetOldest
	saramaConfig.Consumer.Offsets.AutoCommit.Enable = false
	client, err := sarama.NewConsumerGroup(config.KafkaBrokers, config.PodID, saramaConfig)
	if err != nil {
		panic(err)
	}

	consumer := kafka.NewConsumer()

	return &countClicks{
		client:   client,
		consumer: consumer,
		postgres: postgres,
	}
}

func (c *countClicks) Run() {
	go func() {
		for {
			if err := c.client.Consume(context.Background(), []string{config.ClicksTopic}, c.consumer); err != nil {
				time.Sleep(time.Second)
			}
		}
	}()

	channel := c.consumer.GetMessagesChan()
	for message := range channel {
		var clicks models.Clicks
		err := json.Unmarshal(message, &clicks)
		if err != nil {
			log.Printf(err.Error())
			continue
		}

		err = c.postgres.IncClicks(context.Background(), clicks.ShortUrlID, clicks.Inc)
		if err != nil {
			log.Printf(err.Error())
			continue
		}
	}
}
