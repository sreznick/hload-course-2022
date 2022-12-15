package dqueue

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
)

type RedisHost struct {
	Conf redis.Options

	LowerKey int64
	UpperKey int64
}

type RedisCluster struct {
	RedisHosts []RedisHost
}

func (h RedisHost) in_host(key string) bool {
	ctx := context.Background()
	r := redis.NewClient(&h.Conf)

	key_hash, err := r.ClusterKeySlot(ctx, key).Result()
	if err != nil {
		return false
	}

	return key_hash >= h.LowerKey && key_hash <= h.UpperKey
}

func (c RedisCluster) get_key_host(key string) (redis.Options, error) {
	for _, h := range c.RedisHosts {
		if h.in_host(key) {
			return h.Conf, nil
		}
	}

	return redis.Options{}, fmt.Errorf("can't get host that will contain this key")
}
