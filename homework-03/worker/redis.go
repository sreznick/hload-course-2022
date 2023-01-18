package worker

import (
	"github.com/go-redis/redis/v8"
	_ "github.com/go-redis/redis/v8"
)

const (
	prefix = "mdiagilev"
)

type Redis struct {
	Cluster string
	Client  *redis.Client
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
		return err
	}

	client.Client = redisClient
	return nil

}

func (client *Redis) Close() {
	client.Client.Close()
}
