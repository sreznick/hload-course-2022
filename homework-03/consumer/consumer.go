package consumer

import (
	"main/common"
	"main/consumer/kafka"
)

func ConsumerRoutine(c common.KafkaConfig) {
	kafka.SetConsumerKafka(c)
	go KafkaDvij()
	r := SetupWorker()
	err := r.Run(":8081")
	if err != nil {
		panic("Something wrong with router: " + err.Error())
	}
}
