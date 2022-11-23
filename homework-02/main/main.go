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

	hosts := []string{"localhost:6379"}
	dqueue.Config(&redisOptions)
	firstD, _ := dqueue.Open("first", 1, &hosts)
	firstD.Push("bbb")
	firstD.Pull()

}
