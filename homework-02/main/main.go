package main

import (
    "os"
//    "fmt"
    "dqueue"
//    "context"

    "github.com/go-redis/redis/v8"
)

func main() {
    in_bytes, err := os.ReadFile("file.txt")
    if err != nil {
        fmt.Print(err)
        return 1
    }
    in_adresses := strings.Split(string(in_bytes),"\n")
    redisOptions := redis.ClusterOptions {
        Addrs: in_adresses,
    }
    dqueue.Config(&redisOptions)



}
