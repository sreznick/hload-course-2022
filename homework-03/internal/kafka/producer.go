package kafka

import (
	"context"
	"encoding/json"
	"log"

	"github.com/Shopify/sarama"
	"github.com/pkg/errors"
)

type Producer interface {
	Produce(ctx context.Context, key string, data interface{}) error
}

type producer struct {
	producer sarama.AsyncProducer
	topic    string
}

func NewProducer(KafkaBrokers []string, topic string) (*producer, error) {
	cfg := sarama.NewConfig()

	asyncProducer, err := sarama.NewAsyncProducer(KafkaBrokers, cfg)
	if err != nil {
		return nil, errors.Wrap(err, "Failed create kafka producer")
	}

	if err != nil {
		return nil, errors.Wrap(err, "Failed wrap kafka producer")
	}

	return &producer{
		producer: asyncProducer,
		topic:    topic,
	}, nil
}

func (p *producer) Produce(ctx context.Context, key string, data interface{}) error {
	json, err := json.Marshal(data)
	if err != nil {
		return errors.Wrap(err, "Failed marshal message")
	}
	log.Printf("Produce %v\n", string(json))

	msg := sarama.ProducerMessage{
		Topic: p.topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.ByteEncoder(json),
	}

	p.producer.Input() <- &msg

	return nil
}
