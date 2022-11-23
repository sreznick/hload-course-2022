package main

import (
	"dqueue"

	"github.com/go-redis/redis/v8"
)

func main() {
	redisOptions := redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	}

	//hosts := []string{"localhost:7000", "localhost:7001", "localhost:7002", "localhost:7003"}
	hosts := []string{"158.160.9.8:6379", "51.250.106.140:6379", "158.160.19.212:6379", "158.160.19.2:6379"}
	//hosts := []string{"localhost:6379"}
	dqueue.Config(&redisOptions)
	firstD, _ := dqueue.Open("first", 4, &hosts)
	firstD.Push("aaa1")
	firstD.Push("aaa2")
	firstD.Push("aaa3")
	firstD.Push("aaa4")
	firstD.Push("aaa5")

	firstD.Pull()

}
