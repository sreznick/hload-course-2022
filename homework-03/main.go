package main

import (
	"fmt"
	"main/common"
	"main/consumer"
	"main/producer"
	"os"
)

func main() {
	typ := os.Args[1]
	kafkaConfig := common.GetKafkaConfig()

	common.Delim = kafkaConfig.MessageDelim[0]
	common.ClicksThrsh = kafkaConfig.ClicksThrsh

	if typ == "producer" {
		producer.ProducerRoutine(kafkaConfig)
	} else if typ == "consumer" {
		rc := common.GetRedisConfig()
		consumer.ConsumerRoutine(kafkaConfig, rc)
	} else {
		fmt.Println("=(")
	}
}
