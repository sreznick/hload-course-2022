package main

import (
	"fmt"
	_ "github.com/lib/pq"
	_ "github.com/prometheus/client_golang/prometheus"
	"urlShortener/utils"
	_ "urlShortener/utils"

	_ "github.com/confluentinc/confluent-kafka-go/kafka"
)

const key = "bmadzhuga::main"

func main() {
	redis := utils.Redis{Cluster: "158.160.9.8:26379"}
	err := redis.Connect()

	if err != nil {
		panic(err)
	}

	redis.Put("key_c", "val_c")

	response, err := redis.Client.HGetAll(redis.Client.Context(), key).Result()

	if err != nil {
		panic(err)
	}

	for key := range response {
		val := response[key]
		fmt.Printf("{'%v':'%v'}\n", key, val)
	}
}
