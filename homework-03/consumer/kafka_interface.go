package consumer

import (
	"github.com/segmentio/kafka-go"
	"main/common"
)

var UrlsConsumer kafka.Reader

var ClicksProducer kafka.Writer

func SetConsumerKafka(c common.KafkaConfig) {
	UrlsConsumer = *kafka.NewReader(kafka.ReaderConfig{
		Brokers: c.UrlsBrokers,
		GroupID: c.UrlsGroup,
		Topic:   c.UrlTopicName,
	})

	ClicksProducer = kafka.Writer{
		Addr:         kafka.TCP(c.ClicksProducing),
		Topic:        c.ClicksTopicName,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: 1,
	}
}
