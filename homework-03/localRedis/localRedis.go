package localRedis

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"time"
	"urlHandler"

	"github.com/go-redis/redis/v8"
)

type RedisCluster struct {
	RedisOptions redis.Options
	Ctx          context.Context
	Prefix       string
	Master       string
	Workers      []string
}

func CreateRedisCluster(redisOptions redis.Options, ctx context.Context, prefix string, hosts *[]string) RedisCluster {
	return RedisCluster{
		RedisOptions: redisOptions,
		Ctx:          ctx,
		Prefix:       prefix,
		Master:       (*hosts)[0],
		Workers:      (*hosts)[1:len(*hosts)],
	}
}

func GetCurrentId(cluster *RedisCluster) int {
	(*cluster).RedisOptions.Addr = (*cluster).Master
	rdb := redis.NewClient(&(*cluster).RedisOptions)
	result, _ := rdb.Do((*cluster).Ctx, "get", (*cluster).Prefix+"_id").Result()
	var id int

	if result != nil {
		stringId := fmt.Sprintf("%v", result)
		id, _ = strconv.Atoi(stringId)
		id++
	} else {
		id = 0
	}

	rdb.Do((*cluster).Ctx, "set", (*cluster).Prefix+"_id", strconv.Itoa(id))
	return id
}

func GetTinyUrl(cluster *RedisCluster, longUrl string) (string, bool) {
	(*cluster).RedisOptions.Addr = (*cluster).Master
	rdb := redis.NewClient(&(*cluster).RedisOptions)
	result, _ := rdb.Do((*cluster).Ctx, "get", (*cluster).Prefix+"_"+longUrl).Result()
	var tinyUrl string
	var isNew bool
	if result == nil {
		isNew = true
		id := GetCurrentId(cluster)
		tinyUrl = urlHandler.GenerateTinyUrl(id)
		rdb.Do((*cluster).Ctx, "set", (*cluster).Prefix+"_"+longUrl, tinyUrl)
		rdb.Do((*cluster).Ctx, "set", (*cluster).Prefix+"_"+tinyUrl+"_"+"clicks", "0")
	} else {
		isNew = false
		tinyUrl = fmt.Sprintf("%v", result)
	}
	return tinyUrl, isNew
}

func IncreaseClicksCount(cluster *RedisCluster, rdb *redis.Client, tinyUrl string) int {
	clicks := 0
	result, _ := rdb.Do((*cluster).Ctx, "get", (*cluster).Prefix+"_"+tinyUrl+"_"+"clicks").Result()
	if result != nil {
		stringClicks, ok := result.(string)
		if ok {
			clicks, _ = strconv.Atoi(stringClicks)
		}
	}
	clicks++
	rdb.Do((*cluster).Ctx, "set", (*cluster).Prefix+"_"+tinyUrl+"_"+"clicks", strconv.Itoa(clicks))
	return clicks

}
func CheckTinyUrl(cluster *RedisCluster, tinyUrl string) (string, int, error) {
	rand.Seed(time.Now().UnixNano())
	id := rand.Intn(2)

	(*cluster).RedisOptions.Addr = (*cluster).Workers[id]
	rdb := redis.NewClient(&(*cluster).RedisOptions)
	longUrl, _ := rdb.Do((*cluster).Ctx, "get", (*cluster).Prefix+"_"+tinyUrl).Result()
	//fmt.Println(longUrl)
	if longUrl == nil {
		return "", 0, errors.New("no such tiny url")

	}
	clicks := IncreaseClicksCount(cluster, rdb, tinyUrl)
	return fmt.Sprintf("%v", longUrl), clicks, nil

}
