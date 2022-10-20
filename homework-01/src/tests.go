package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

type request struct {
	Longurl string
}

type put struct {
	Longurl string
	Tinyurl string
}

func makePut() string {
	body := request{
		Longurl: "https://github.com/ACE-777",
	}
	data, err := json.Marshal(body)
	if err != nil {
		fmt.Println(err)
		panic("error")
	}
	client := &http.Client{}

	request, err := http.NewRequest(http.MethodPut, "http://localhost:8080/create", bytes.NewBuffer(data))
	if err != nil {
		fmt.Println(err)
		panic("error")
	}
	request.Header.Set("Content-Type", "application/json; charset=utf-8")
	defer request.Body.Close()

	response, err := client.Do(request)
	if err != nil {
		fmt.Println(err)
		panic("error")
	}
	defer response.Body.Close()

	var responseForm put
	err = json.NewDecoder(response.Body).Decode(&responseForm)
	if err != nil {
		fmt.Println(err)
		panic("error")
	}

	return responseForm.Tinyurl

}
func makeGet(tinyUrl string) {
	response, err := http.Get("http://localhost:8080/" + tinyUrl)
	if response != nil {
		response.Body.Close()
	}
	if err != nil {
		fmt.Println(err)
		panic("error")
	}
}

func main() {
	tinyUrl := makePut()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		for k := 1; k <= 10000; k++ {
			makePut()

			if k == 10000 {
				fmt.Println("Test on PUT request was completed")
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for k := 1; k <= 100000; k++ {
			makeGet(tinyUrl)

			if k == 100000 {
				fmt.Println("Test on valid GET request was completed")
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for k := 1; k <= 100000; k++ {
			makeGet("invalidlinkq1w2e3r4t5y6")

			if k == 100000 {
				fmt.Println("Test on invalid GET request was completed")
			}
		}
	}()

	wg.Wait()

}
