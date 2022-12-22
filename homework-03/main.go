package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"localKafka"
	"localRedis"
	"strings"

	"github.com/go-redis/redis/v8"
	"main.go/server"
)

func readHostsFromFile(filename string) []string {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println(err)
	}
	dataString := string(data)
	return strings.Split(dataString, " ")
}

func main() {
	redisOptions := redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	}

	redisHosts := readHostsFromFile("redisHosts")
	redisCluster := localRedis.CreateRedisCluster(redisOptions, context.Background(), "aisakova", &redisHosts)

	urlWriter := localKafka.CreateUrlWriter()
	urlReaders := localKafka.CreateUrlReaders(len(redisCluster.Workers))
	//fmt.Println(redisCluster)

	r := server.SetupRouter(&redisCluster, urlWriter, urlReaders)
	/*longUrl := "https://hey"
	tinyUrl := localRedis.GetTinyUrl(&redisCluster, longUrl)
	fmt.Println(tinyUrl)*/
	r.Run("0.0.0.0:26397")
	//fmt.Println(aaa.a())

}
