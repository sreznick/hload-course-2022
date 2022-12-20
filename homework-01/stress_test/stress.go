package main

import (
	"bytes"
	"fmt"
	"net/http"
)

var createRequestNumber = 1

func makeCreateRequest() {
	createRequestNumber++
	bodyJSON := []byte(fmt.Sprintf(`{"longurl": "http://yandex.ru/%d"}`, createRequestNumber))
	request, err := http.NewRequest("PUT", "http://localhost:8080/create", bytes.NewBuffer(bodyJSON))
	if err != nil {
		panic(err)
	}
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")

	client := &http.Client{}
	res, err := client.Do(request)

	if err != nil {
		panic(err)
	}

	if res.StatusCode != 200 {
		panic(res.StatusCode)
	}
}

func makeSuccessfulGetUrlRequest() {
	res, err := http.Get("http://localhost:8080/caaaaaa")

	if err != nil {
		panic(err)
	}

	if res.StatusCode != 200 {
		panic(res.StatusCode)
	}
}

func makeUnsuccessfulGetUrlRequest() {
	res, err := http.Get("http://localhost:8080/weirdoo")

	if err != nil {
		panic(err)
	}

	if res.StatusCode != 404 {
		panic(res.StatusCode)
	}
}

func main() {
	const createCount = 1 // 10000
	const unsuccessfulGetCount = 100000
	const successfulGetCount = 100000

	for i := 1; i <= createCount; i++ {
		makeCreateRequest()
		if i%1000 == 0 {
			fmt.Println("Create ", i)
		}
	}

	for i := 1; i <= successfulGetCount; i++ {
		makeSuccessfulGetUrlRequest()
		if i%100 == 0 {
			fmt.Println("Get successful", i)
		}
	}

	for i := 1; i <= unsuccessfulGetCount; i++ {
		makeUnsuccessfulGetUrlRequest()
		if i%1000 == 0 {
			fmt.Println("Get unsuccessful", i)
		}
	}
}
