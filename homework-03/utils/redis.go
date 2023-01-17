package utils

import (
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	_ "github.com/go-redis/redis/v8"
)

const (
	prefix = "bmadzhuga"
)

type Redis struct {
	Cluster string
	Client  *redis.Client
	Name    string
}

func (client *Redis) GetDefaultKey() string {
	name := client.Name
	if name == "" {
		name = "name"
	}
	return client.GetKey(name)
}

func (client *Redis) GetKey(name string) string {
	return fmt.Sprintf("%s::%s", prefix, name)
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

func (client *Redis) Push(value string) error {
	if client.Client == nil {
		return errors.New("Empty redis client")
	}

	client.Client.RPush(client.Client.Context(), client.GetDefaultKey(), value)

	return nil
}

func (client *Redis) Put(field string, val string) error {
	return client.PutWithKey(client.GetDefaultKey(), field, val)
}

func (client *Redis) PutWithKey(key string, field string, val string) error {
	if client.Client == nil {
		return errors.New("Empty redis client")
	}

	err := client.Client.HSet(client.Client.Context(), client.GetKey(key), field, val)

	if err != nil {
		return err.Err()
	}

	return nil
}

func (client *Redis) Pull() (string, error) {
	if client.Client == nil {
		return "", errors.New("Empty redis client")
	}

	res, err := client.Client.LPop(client.Client.Context(), client.GetDefaultKey()).Result()

	if err != nil {
		return "", err
	}

	return res, nil
}

func (client *Redis) GetMap() (map[string]string, error) {
	return client.GetMapWithKey(client.GetDefaultKey())
}

func (client *Redis) GetMapWithKey(key string) (map[string]string, error) {

	if client == nil || client.Client == nil {
		return nil, errors.New("Empty Redis Client")
	}
	return client.Client.HGetAll(client.Client.Context(), client.GetKey(key)).Result()
}

func (client *Redis) Close() {
	client.Client.Close()
}
