package main

import (
//    "fmt"
    "dqueue"
//    "context"

    "github.com/go-redis/redis/v8"
)

func main() {
    redisOptions := redis.Options{
        Addr:     "localhost:6379",
        Password: "", // no password set
        DB:       0,  // use default DB
    }

    dqueue.Config(&redisOptions, []string{"127.0.0.1"})
}
