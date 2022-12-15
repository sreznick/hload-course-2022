package main

import (
	"dqueue"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"sync"
)

func test_queue(name string) {
	q, err := dqueue.Open(name, 2)
	if err != nil {
		fmt.Println(name)
		panic(err)
	}

	var pushed []string
	for i := 0; i < 3; i++ {
		msg := "Hi " + strconv.Itoa(i)
		err = q.Push(msg)
		if err != nil {
			panic(err)
		}
		pushed = append(pushed, msg)
	}

	var pulled []string
	for i := 0; i < 3; i++ {
		v, err := q.Pull()
		if err != nil {
			panic(err)
		}
		pulled = append(pulled, v)
	}

	sort.Strings(pulled)
	sort.Strings(pushed)
	if !reflect.DeepEqual(pulled, pushed) {
		panic(fmt.Sprintf("!ACHTUNG! %s\n", name))
	}
}

func test_one_queue(name string) {
	q, err := dqueue.Open(name, 4)
	if err != nil {
		fmt.Println(name)
		panic(err)
	}

	for i := 0; i < 3; i++ {
		msg := "Hi " + strconv.Itoa(i)
		err = q.Push(msg)
		if err != nil {
			panic(err)
		}
	}

	for i := 0; i < 3; i++ {
		_, err := q.Pull()
		if err != nil {
			panic(err)
		}
	}
}

func main() {
	r, z, err := ReadConfig()
	dqueue.Config(r, z)
	if err != nil {
		panic(err)
	}

	var wg sync.WaitGroup
	{
		for i := 0; i < 6; i++ {
			wg.Add(1)
			i_ := i + 1
			go func() {
				test_queue("small_queue" + strconv.Itoa(i_))
				wg.Done()
			}()
		}

		wg.Wait()
		fmt.Println("Group 1 done")
	}

	{
		for i := 0; i < 6; i++ {
			wg.Add(1)
			go func() {
				test_one_queue("big_queue")
				wg.Done()
			}()
		}

		wg.Wait()
		fmt.Println("Group 2 done")
	}

}
