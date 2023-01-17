package utils

import (
	"errors"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
)

const (
	broker  = "158.160.19.212:9092"
	group   = "bmadzhuga-consumer-group"
	timeout = 6000
)

type Kafka struct {
	Topic    string
	Type     string
	Producer *kafka.Producer
	Consumer *kafka.Consumer
}

func (client *Kafka) Connect() error {
	kafkaProducer, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": broker, "acks": 1})

	if err != nil {
		return err
	}

	client.Producer = kafkaProducer

	kafkaConsumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":  broker,
		"group.id":           group,
		"auto.offset.reset":  "latest",
		"enable.auto.commit": false,
	})

	if err != nil {
		return err
	}

	err = kafkaConsumer.Subscribe(client.Topic, nil)

	if err != nil {
		return err
	}

	client.Consumer = kafkaConsumer

	return nil
}

func (client *Kafka) Send(longUrl string, tinyUrl string, status bool) error {

	statusString := "success"

	if !status {
		statusString = "failed"
	}

	var err = client.Producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &client.Topic, Partition: kafka.PartitionAny},
		Value:          []byte(longUrl + "::" + tinyUrl + "::" + statusString)},
		nil,
	)

	if err != nil {
		errors.New("Can't send message")
	}

	fmt.Println("Message send to topic: %v", client.Topic)

	return nil
}

func (client *Kafka) ReadFromTopic() (string, error) {
	if client.Consumer == nil {
		return "", errors.New("Empty consumer")
	}

	fmt.Println("Start reading from topic: %v", client.Topic)

	for {
		msg, err := client.Consumer.ReadMessage(-1)
		if err == nil {
			return string(msg.Value), nil
		} else {
			return "", err
		}
	}

	return "", errors.New("Error while reading")
}

func (client *Kafka) Close() {
	client.Producer.Close()
	client.Consumer.Close()
}
