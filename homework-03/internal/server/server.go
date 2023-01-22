package server

import (
	"main/internal/kafka"
	"main/internal/postgres"
	"main/internal/redis"
	masterServer "main/internal/server/master_server"
	simpleServer "main/internal/server/simple_server"
)

type Interface interface {
	Run()
}

func NewSimpleServer(producer kafka.Producer, redis redis.Interface) Interface {
	return simpleServer.New(producer, redis)
}

func NewMasterServer(producer kafka.Producer, postgres postgres.Interface) Interface {
	return masterServer.New(producer, postgres)
}
