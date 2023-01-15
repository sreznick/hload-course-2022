package producer

import (
	"github.com/segmentio/kafka-go"
	"main/common"
)

var urlsProducer kafka.Writer

var clicksReader kafka.Reader

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
