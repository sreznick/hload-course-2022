package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/go-redis/redis/v8"
)

type RedisCluster struct {
	redisOptions redis.Options
	ctx          context.Context
	prefix       string
	master       string
	workers      []string
}

func readHostsFromFile(filename string) []string {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println(err)
	}
	dataString := string(data)
	return strings.Split(dataString, " ")
}

func createRedisCluster(redisOptions redis.Options, ctx context.Context, prefix string, hosts *[]string) RedisCluster {
	return RedisCluster{
		redisOptions: redisOptions,
		ctx:          ctx,
		prefix:       prefix,
		master:       (*hosts)[0],
		workers:      (*hosts)[1:len(*hosts)],
	}

}

func main() {
	redisOptions := redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	}
	hosts := readHostsFromFile("redisHosts")
	redisCluster := createRedisCluster(redisOptions, context.Background(), "aisakova", &hosts)
	fmt.Println(redisCluster)
	//fmt.Println(hosts[1])
}
