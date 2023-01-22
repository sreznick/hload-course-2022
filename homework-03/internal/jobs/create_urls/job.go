package createurls

import (
	"context"
	"encoding/json"
	"log"
	"main/internal/config"
	"main/internal/kafka"
	"main/internal/models"
	"main/internal/redis"
	"time"

	"github.com/Shopify/sarama"
)

type createUrls struct {
	client   sarama.ConsumerGroup
	consumer kafka.Consumer
	redis    redis.Interface
}

func NewCreateUrlsJob(redis redis.Interface) *createUrls {
	saramaConfig := sarama.NewConfig()
	saramaConfig.Consumer.Offsets.Initial = sarama.OffsetOldest
	saramaConfig.Consumer.Offsets.AutoCommit.Enable = false
	client, err := sarama.NewConsumerGroup(config.KafkaBrokers, config.PodID, saramaConfig)
	if err != nil {
		panic(err)
	}

	consumer := kafka.NewConsumer()

	return &createUrls{
		client:   client,
		consumer: consumer,
		redis:    redis,
	}
}

func (c *createUrls) Run() {
	go func() {
		for {
			if err := c.client.Consume(context.Background(), []string{config.CreateTopic}, c.consumer); err != nil {
				time.Sleep(time.Second)
			}
		}
	}()

	channel := c.consumer.GetMessagesChan()
	for message := range channel {
		var create models.Create

		err := json.Unmarshal(message, &create)
		if err != nil {
			log.Printf(err.Error())
			continue
		}

		c.redis.AddUrl(context.Background(), create.LongUrl, create.ShortUrlID)
	}
}
