package main

import (
	"dqueue"
	"fmt"
	"io/ioutil"
	"strings"
)

/*
Read files .redisAddr and .zkAddr than split it by \n.
*/
func readConfig() ([]string, []string) {
	contentRedis, err := ioutil.ReadFile(".redisAddr")
	if err != nil {
		panic(err)
	}
	contentZk, err := ioutil.ReadFile(".zkAddr")
	if err != nil {
		panic(err)
	}
	return strings.Split(string(contentRedis), "\n"), strings.Split(string(contentZk), "\n")
}

func main() {
	redisAddr, zkAddr := readConfig()
	dqueue.Config(redisAddr, zkAddr)

	dq, err := dqueue.Open("test", 1)
	if err != nil {
		fmt.Println(err)
		return
	}
	errPush := dq.Push("123")
	if errPush != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Push %d\n", 123)
	errPush = dq.Push("456")
	if errPush != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Push %d\n", 789)
	errPush = dq.Push("789")
	if errPush != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Push %d\n", 465)

	pullData, err := dq.Pull()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Pull %s\n", pullData)
	pullData, err = dq.Pull()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Pull %s\n", pullData)
}
