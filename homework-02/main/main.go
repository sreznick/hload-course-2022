package main

import (
	"dqueue"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/go-redis/redis/v8"
)

func main() {
	redisOptions := redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	}

	data, err := ioutil.ReadFile("hosts.txt")
	if err != nil {
		fmt.Println(err)
	}
	dataString := string(data)
	hosts := strings.Split(dataString, " ")

	dqueue.Config(&redisOptions)
	firstD, _ := dqueue.Open("first", len(hosts), &hosts)
	firstD.Push("aaa1")
	firstD.Push("aaa2")
	firstD.Push("aaa3")
	firstD.Push("aaa4")
	firstD.Push("aaa5")

	firstD.Pull()

}
