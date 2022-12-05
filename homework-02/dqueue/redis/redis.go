package redis

import (
	"context"
	"dqueue/models"
	"encoding/json"

	redis "github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
)

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"

var sufs []string

func init() {
	for _, a := range letters {
		for _, b := range letters {
			for _, c := range letters {
				sufs = append(sufs, string([]rune{a, b, c}))
			}
		}
	}
}

type redisClient struct {
	options *redis.ClusterOptions
}

func NewRedisClient() *redisClient {
	return &redisClient{}
}

func (c *redisClient) Configure(redisOptions *redis.ClusterOptions) {
	c.options = redisOptions
}

func (c *redisClient) Create(name string, n int) ([]*models.Host, error) {
	client := redis.NewClusterClient(c.options)
	var result []*models.Host

	if n > 4 {
		return nil, errors.New("Too much shards")
	}
	used := make(map[string]bool)
	for _, addr := range c.options.Addrs {
		used[addr] = false
	}

	clusterSlots, err := client.ClusterSlots(context.Background()).Result()
	if err != nil {
		return nil, errors.New("Internal Error")
	}

	for _, suf := range sufs {
		slot, err := client.ClusterKeySlot(context.Background(), name+":"+suf).Result()
		if err != nil {
			continue
		}

		for _, clusterSlot := range clusterSlots {
			if int64(clusterSlot.Start) <= slot && slot <= int64(clusterSlot.End) {
				if use, ok := used[clusterSlot.Nodes[0].Addr]; ok && !use {
					used[clusterSlot.Nodes[0].Addr] = true
					result = append(result, &models.Host{
						Addr: clusterSlot.Nodes[0].Addr,
						Key:  name + ":" + suf,
					})
				}
			}
		}

		if len(result) == n {
			break
		}
	}

	if len(result) < n {
		return nil, errors.New("Internal Error, shards not selected")
	}
	return result, nil
}

func (c *redisClient) Front(host *models.Host) (*models.RedisValue, error) {
	client := redis.NewClient(&redis.Options{Addr: host.Addr, Password: "", DB: 0})
	value, err := client.LIndex(context.Background(), host.Key, 0).Result()

	if err != nil {
		return nil, err
	}

	var result models.RedisValue
	err = json.Unmarshal([]byte(value), &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *redisClient) Pop(host models.Host) error {
	client := redis.NewClient(&redis.Options{Addr: host.Addr, Password: "", DB: 0})

	_, err := client.LPop(context.Background(), host.Key).Result()
	return err
}

func (c *redisClient) Push(host models.Host, value models.RedisValue) error {
	client := redis.NewClient(&redis.Options{Addr: host.Addr, Password: "", DB: 0})

	data, err := json.Marshal(value)
	if err != nil {
		return errors.Wrap(err, "Internal error")
	}

	_, err = client.RPush(context.Background(), host.Key, string(data)).Result()
	return err
}
