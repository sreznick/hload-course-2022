package consumer

import (
	"main/common"
)

func ConsumerRoutine(c common.KafkaConfig, rc common.RedisConfig) {
	SetConsumerKafka(c)
	SetRedisOpts(rc)
	go ReadUrls()
	r := SetupWorker()
	err := r.Run(":8081")
	if err != nil {
		panic("Something wrong with router: " + err.Error())
	}
}
