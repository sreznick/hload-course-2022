package main

import (
	"dqueue"
	"fmt"
	"math/rand"
	"strconv"
	"time"
)

func main() {

	redis := []string{"158.160.9.8:6379", "51.250.106.140:6379", "158.160.19.212:6379", "158.160.19.2:6379"}
	zookeeper := []string{"158.160.9.8", "51.250.106.140", "158.160.19.212", "158.160.19.2"}

	dqueue.Config(redis, zookeeper)

	rand.Seed(time.Now().UnixNano())

	dqueue, err := dqueue.Open("main", 4)
	if err != nil {
		fmt.Println(err)
		return
	}

	for i := 0; i < 100; i++ {
		data := strconv.Itoa(rand.Intn(100))
		err := dqueue.Push(data)
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	for i := 0; i < 100; i++ {
		data, err := dqueue.Pull()
		if err != nil {
			fmt.Println(err)
			return
		} else {
			fmt.Println(data)
		}
	}

}
