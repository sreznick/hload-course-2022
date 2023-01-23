package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

const URL = "http://localhost:8080/"

type PutRequestBody struct {
	Longurl string
}

type PutResponseBody struct {
	Longurl string
	Tinyurl string
}

func checkError(err error, msg string) {
	if err != nil {
		fmt.Println(err)
		panic(msg)
	}
}
func createPutRequest() string {
	body := PutRequestBody{
		Longurl: "https://emkn.ru",
	}
	data, err := json.Marshal(body)
	checkError(err, "Could not create json")
	client := &http.Client{}

	request, err := http.NewRequest(http.MethodPut, URL+"create", bytes.NewBuffer(data))
	checkError(err, "Could not create http PUT request")
	request.Header.Set("Content-Type", "application/json; charset=utf-8")

	response, err := client.Do(request)
	checkError(err, "Could not do http PUT request")
	defer response.Body.Close()

	var responseJson PutResponseBody
	err = json.NewDecoder(response.Body).Decode(&responseJson)
	checkError(err, "Could not parse json response")

	return responseJson.Tinyurl

}
func createGetRequest(tinyUrl string) {
	_, err := http.Get(URL + tinyUrl)
	checkError(err, "Could not do http GET request")
}

func main() {
	tinyUrl := createPutRequest()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 1; i <= 10000; i++ {
			createPutRequest()

			if i == 10000 {
				fmt.Println("PUT done")
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 1; i <= 100000; i++ {
			createGetRequest(tinyUrl)

			if i == 100000 {
				fmt.Println("Successful GET done")
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 1; i <= 100000; i++ {
			createGetRequest("abracadabra")

			if i == 100000 {
				fmt.Println("Failed GET done")
			}
		}
	}()

	wg.Wait()

}
