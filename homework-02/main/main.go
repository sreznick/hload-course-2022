package main

import (
	"dqueue"

	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/go-zookeeper/zk"
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
	zkHosts := readHostsFromFile("zkHosts")
	c, _, err := zk.Connect(zkHosts, time.Second) //*10)
	if err != nil {
		panic(err)
	}
	_, err = c.Create("/zookeeper/aisakova", []byte("mainFolder"), 0, zk.WorldACL(zk.PermAll))
	if err == nil || err.Error() == "zk: node already exists" {
		fmt.Println("Successfully connected ZooKeeper")
	}

	redisOptions := redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	}

	redisHosts := readHostsFromFile("redisHosts")

	dqueue.Config(&redisOptions)
	firstD, err := dqueue.Open("second", len(redisHosts), &redisHosts, c)

	if err == nil {
		firstD.Push("aaa1")
		firstD.Push("aaa2")
		firstD.Push("aaa3")
		firstD.Push("aaa4")
		firstD.Push("aaa5")

		firstD.Pull()
	}

}
