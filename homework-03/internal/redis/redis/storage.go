package redis

import (
	"fmt"
	"log"
	"main/internal/config"
	"strconv"

	"github.com/go-redis/redis/v9"
	"golang.org/x/net/context"
)

type client struct {
	rdb *redis.Client
}

func New(address string) *client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     address,
		Password: "",
		DB:       0,
	})

	return &client{
		rdb: rdb,
	}
}

func (r *client) GetUrl(ctx context.Context, shortUrlID int64) (string, error) {
	log.Printf("GetUrl %v\n", shortUrlID)

	res := r.rdb.Get(ctx, fmt.Sprintf("%surl:%d", config.RedisPrefix, shortUrlID))
	if res.Err() != nil {
		return "", res.Err()
	}

	return res.Val(), nil
}

func (r *client) AddUrl(ctx context.Context, longUrl string, shortUrlID int64) error {
	log.Printf("AddUrl %v %v\n", longUrl, shortUrlID)

	return r.rdb.Set(ctx, fmt.Sprintf("%surl:%d", config.RedisPrefix, shortUrlID), longUrl, 0).Err()
}

func (r *client) ChangeClicks(ctx context.Context, shortUrlID, cnt int64) (int64, error) {
	log.Printf("ChangeClicks %v %v\n", cnt, shortUrlID)

	err := r.rdb.IncrBy(ctx, fmt.Sprintf("%sclicks:%d", config.RedisPrefix, shortUrlID), cnt).Err()
	if err != nil {
		return 0, err
	}

	res := r.rdb.Get(ctx, fmt.Sprintf("%sclicks:%d", config.RedisPrefix, shortUrlID))
	if res.Err() != nil {
		return 0, res.Err()
	}

	clicks, err := strconv.ParseInt(res.Val(), 10, 64)
	if err != nil {
		return 0, err
	}

	return clicks, nil
}
