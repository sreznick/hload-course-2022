package worker

import (
	"fmt"
	"github.com/go-redis/redis/v8"
	_ "github.com/go-redis/redis/v8"
)

const (
	prefix = "mdiagilev"
)

type Redis struct {
	*redis.Client

	Cluster string
	Name    string
}

func (client *Redis) Connect() error {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     client.Cluster,
		Password: "",
		DB:       0,
	})

	_, err := redisClient.Ping(redisClient.Context()).Result()
	if err != nil {
		return fmt.Errorf("redisClient.Ping: %w", err)
	}

	client.Client = redisClient
	return nil
}
