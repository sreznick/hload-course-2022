package main

import (
	"main/internal/config"
	countClicks "main/internal/jobs/count_clicks"
	createUrls "main/internal/jobs/create_urls"
	"main/internal/kafka"
	"main/internal/postgres"
	"main/internal/redis"
	"main/internal/server"
)

func startMaster() {
	postgres, err := postgres.New()
	if err != nil {
		panic(err)
	}

	producer, err := kafka.NewProducer(config.KafkaBrokers, config.CreateTopic)
	if err != nil {
		panic(err)
	}

	countClicksJob := countClicks.NewCountClicksJob(postgres)
	go countClicksJob.Run()

	server := server.NewMasterServer(producer, postgres)
	server.Run()
}

func start() {
	redis, err := redis.New(config.RedisAddr)
	if err != nil {
		panic(err)
	}

	producer, err := kafka.NewProducer(config.KafkaBrokers, config.ClicksTopic)
	if err != nil {
		panic(err)
	}

	createUrlsJob := createUrls.NewCreateUrlsJob(redis)
	go createUrlsJob.Run()

	server := server.NewSimpleServer(producer, redis)
	server.Run()
}

func main() {
	if config.IsMaster {
		startMaster()
	} else {
		start()
	}
}
