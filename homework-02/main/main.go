package main

import (
	"dqueue"
	"os"

	"fmt"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/go-zookeeper/zk"
)

func readHosts(filename string) []string {
	fileData, err := os.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	return strings.Split(string(fileData), " ")
}

const connectPath = "/zookeeper/epostnikov"

func initQueue() (dqueue.DQueue, error) {
	dqueue.Config(&redis.Options{
		Password:   "",
		DB:         0,
		MaxRetries: 10,
	})

	c, _, err := zk.Connect(readHosts("../zk"), time.Second*3)
	if err != nil {
		panic(err)
	}
	_, err = c.Create(connectPath, []byte("mainFolder"), 0, zk.WorldACL(zk.PermAll))
	if err == nil || err.Error() == "zk: node already exists" {
		fmt.Println("ZooKeeper is connected")
	}

	redisHosts := readHosts("../redis")
	return dqueue.Open("epostnikov2", len(redisHosts), &redisHosts, c)
}

func main() {
	que, err := initQueue()

	que.Clear()
	if err == nil {
		_ = que.Push("1")
		_ = que.Push("2")
		_ = que.Push("3")

		v1, e1 := que.Pull()
		v2, e2 := que.Pull()
		v3, e3 := que.Pull()
		fmt.Printf("errors %s\n %s\n %s\n", e1, e2, e3)
		fmt.Printf("Executed fine if %s %s %s is 1 2 3", v1, v2, v3)
	} else {
		print("Has error", err)
	}

}
