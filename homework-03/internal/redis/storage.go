package redis

import (
	"context"
	"main/internal/redis/redis"
)

type Interface interface {
	GetUrl(ctx context.Context, id int64) (string, error)
	AddUrl(ctx context.Context, longUrl string, shortUrlID int64) error
	ChangeClicks(ctx context.Context, shortUrlID, cnt int64) (int64, error)
}

func New(address string) (Interface, error) {
	return redis.New(address), nil
}
