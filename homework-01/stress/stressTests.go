package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type Request struct {
	Long_url string `json:"longurl"`
}

type Response struct {
	Long_url string `json:"longurl"`
	Tiny_url string `json:"tinyurl"`
}

func ErrorCheck(err error, message string) {
	if err != nil {
		fmt.Println(message, err)
		panic("exit")
	}
}

func doCreateRequest(LongUrlInput string) string {
	body := Request{
		Long_url: LongUrlInput,
	}

	data, err := json.Marshal(body)
	ErrorCheck(err, "ERROR: Cannot create json.")
	client := &http.Client{}

	request, err := http.NewRequest(http.MethodPut, URL+"create", bytes.NewBuffer(data))
	ErrorCheck(err, "ERROR: Could not create http PUT request")
	request.Header.Set("Content-Type", "application/json; charset=utf-8")

	response, err := client.Do(request)
	ErrorCheck(err, "ERROR: Cannot do http PUT request.")
	defer response.Body.Close()

	var responseAsJson Response
	err = json.NewDecoder(response.Body).Decode(&responseAsJson)
	ErrorCheck(err, "ERROR: Cannot parse json response.")

	return responseAsJson.Tiny_url
}

const URL = "http://localhost:8080/"

func doGetRequest(tinyUrl string, client *http.Client) int {
	resp, err := client.Get(URL + tinyUrl)
	ErrorCheck(err, "Could not do http GET request")
	defer resp.Body.Close()
	return resp.StatusCode
}

const (
	redirectCode = 302
	notFoundCode = 404
)

func main() {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	url := "https://github.com/sreznick/hload-course-2022/tree/master/homework-01"
	tinyUrl := doCreateRequest(url)

	for i := 1; i <= 10000; i++ {
		fmt.Println(i)
		doCreateRequest(url + fmt.Sprint(i))
	}

	for i := 1; i <= 100000; i++ {
		code := doGetRequest(tinyUrl, client)
		if code != redirectCode {
			panic("Code is not 302")
		}
	}

	for i := 1; i <= 100000; i++ {
		code := doGetRequest(tinyUrl+"jrbvrjbqevqrklqv", client)
		if code != notFoundCode {
			panic("Code is not 404")
		}
	}
}
