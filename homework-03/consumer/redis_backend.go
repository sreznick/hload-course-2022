package consumer

import (
	"context"
	"strconv"

	"github.com/go-redis/redis/v8"
)

var redisOpts redis.Options = redis.Options{
	Addr:     "",
	Password: "", // no password set
	DB:       0,  // use default DB
}

func buildKey(tinyUrl string) string {
	return "dkulikov:urls:" + tinyUrl
}

func buildClicksKey(tinyUrl string) string {
	return buildKey(tinyUrl) + ":clicks"
}

func GetLongUrl(tinyUrl string) (string, error) {
	ctx := context.Background()
	r := redis.NewClient(&redisOpts)
	return r.Get(ctx, buildKey(tinyUrl)).Result()
}

func SetLongUrl(tinyUrl string, longUrl string) error {
	ctx := context.Background()
	r := redis.NewClient(&redisOpts)

	err := r.Set(ctx, buildKey(tinyUrl), longUrl, 0).Err()
	return err
}

func CreateClick(tinyUrl string) error {
	ctx := context.Background()
	r := redis.NewClient(&redisOpts)

	err := r.Set(ctx, buildClicksKey(tinyUrl), 0, 0).Err()
	return err
}

func IncrementClick(tinyUrl string) error {
	ctx := context.Background()
	r := redis.NewClient(&redisOpts)

	err := r.Incr(ctx, buildClicksKey(tinyUrl)).Err()
	return err
}

func GetClicks(tinyUrl string) (int, error) {
	ctx := context.Background()
	r := redis.NewClient(&redisOpts)
	c, err := r.Get(ctx, buildClicksKey(tinyUrl)).Result()
	if err != nil {
		return 0, err
	}

	ci, err := strconv.Atoi(c)
	return ci, err
}
