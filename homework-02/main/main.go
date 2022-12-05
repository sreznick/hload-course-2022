package main

import (
	"dqueue"

	redis "github.com/go-redis/redis/v8"
)

func main() {
	redisOptions := redis.ClusterOptions{
		Addrs:    []string{"localhost:6379"},
		Password: "",
	}

	zkCluster := []string{"localhost"}

	dqueue.Config(&redisOptions, zkCluster)
}
