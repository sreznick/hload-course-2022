package main

import (
	"dqueue"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/go-redis/redis/v8"
)

func tilt(err error) {
	if err != nil {
		panic(err)
	}
}

func readAddr() ([]string, []string) {
	redisAddr, err := ioutil.ReadFile(".redisAddr")
	tilt(err)
	zkAddr, err := ioutil.ReadFile(".zkAddr")
	tilt(err)
	return strings.Split(string(zkAddr), "\n"), strings.Split(string(redisAddr), "\n")
}

func main() {
	redisOptions := redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	}
	zkAddr, redisAddr := readAddr()

	dqueue.Config(&redisOptions, zkAddr, redisAddr)

	q1, err := dqueue.Open("q1", 4)
	tilt(err)
	err = q1.Push("val")
	tilt(err)
	val, err := q1.Pull()
	tilt(err)
	fmt.Println(val)
}
