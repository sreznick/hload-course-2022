package main

import (
	"bytes"
	"encoding/json"
	"main/server"
	"net/http"
	"sync"
)

const (
	URL                              = "http://localhost:8080/"
	URL_TO_SHORTEN                   = "https://ya.ru"
	CREATE_RESPONSE_NUMBER           = 10_000
	SUCCESSFULL_GET_RESPONSES_NUMBER = 100_000
	FAILED_GET_RESPONSES_NUMBER      = 100_000
)

type CreateRequestBody struct {
	Longurl string
}

type CreateResponseBody struct {
	Longurl string
	Tinyurl string
}

func createRequest() string {
	body := CreateRequestBody{
		Longurl: URL_TO_SHORTEN,
	}
	data, err := json.Marshal(body)
	client := &http.Client{}

	request, err := http.NewRequest(http.MethodPut, URL+"create", bytes.NewBuffer(data))
	request.Header.Set("Content-Type", "application/json; charset=utf-8")
	response, err := client.Do(request)
	defer response.Body.Close()

	var responseJson CreateResponseBody
	err = json.NewDecoder(response.Body).Decode(&responseJson)
	server.HandleError(err, "Could not parse response")

	return responseJson.Tinyurl

}
func createGetRequest(tinyUrl string) {
	_, err := http.Get(URL + tinyUrl)
	server.HandleError(err, "Connection error")
}

func main() {
	tinyUrl := createRequest()
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 1; i <= CREATE_RESPONSE_NUMBER; i++ {
			createRequest()
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 1; i <= SUCCESSFULL_GET_RESPONSES_NUMBER; i++ {
			createGetRequest(tinyUrl)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 1; i <= FAILED_GET_RESPONSES_NUMBER; i++ {
			createGetRequest("abracadabra")
		}
	}()
	wg.Wait()
}
