package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

const URL = "http://localhost:8080/"
const testCnt = 10

type PutRequestJsonBody struct {
	Longurl string
}

type PutResponseJsonBody struct {
	Longurl string
	Tinyurl string
}

var tinyLongUrls map[string]string = make(map[string]string)

func createPutRequest(longurl string) string {
	var jsonBody = []byte(fmt.Sprintf("{\"longurl\": \"%s\"}", longurl))

	request, err := http.NewRequest("PUT", URL+"create", bytes.NewBuffer(jsonBody))
	if err != nil {
		panic("Could not create http PUT request" + err.Error())
	}

	request.Header.Set("Content-Type", "application/json; charset=utf-8")
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		panic("Could not do http PUT request" + err.Error())
	}

	defer response.Body.Close()

	var responseJson PutResponseJsonBody
	err = json.NewDecoder(response.Body).Decode(&responseJson)
	if err != nil {
		panic("Could not parse json response" + err.Error())
	}

	return responseJson.Tinyurl
}

func genLongUrl(i int) string {
	url := "https://google.com/"
	postfix := fmt.Sprintf("%06d", i)
	return url + postfix
}

func testPut() {
	println("TestPut is started")
	for i := 0; i < testCnt; i++ {
		longUrl := genLongUrl(i)
		tinyurl := createPutRequest(longUrl)
		tinyLongUrls[tinyurl] = longUrl
	}
	println("TestPut is passed")
}

func testGetGood() {
	println("TestGetGood is started")
	for tinyUrl, _ := range tinyLongUrls {
		client := &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}

		request, err := http.NewRequest("GET", URL+tinyUrl, nil)
		if err != nil {
			panic("Could not create http GET request" + err.Error())
		}

		response, err := client.Do(request)
		if err != nil {
			panic("Could not do http GET request" + err.Error())
		}

		if response.StatusCode != 302 {
			panic("StatusCode should be 302")
		}
	}
	println("TestGetGood is passed")
}

func genBadTinyUrl(i int) string {
	body := fmt.Sprintf("%05d", i)
	return "Z" + body + "Z" // такой шортурл будет занят только после более 10^6 операций
}

func testGetBad() {
	println("TestGetBad is started")

	for i := 0; i < testCnt; i++ {
		client := &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}

		tinyUrl := genBadTinyUrl(i)
		request, err := http.NewRequest("GET", URL+tinyUrl, nil)
		if err != nil {
			panic("Could not create http GET request" + err.Error())
		}

		response, err := client.Do(request)
		if err != nil {
			panic("Could not do http GET request" + err.Error())
		}

		if response.StatusCode != 404 {
			panic("StatusCode should be 404")
		}
	}
	println("TestGetBad is passed")
}

func main() {
	testPut()
	testGetGood()
	testGetBad()
}
