package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"
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

func createPutRequest(longurl string) string {
	var jsonBody = []byte(`{
		"longurl": "aboba"
	}`)

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

func testPut() {
	url := "abacabadabacaba"
	tinyurl_prev := createPutRequest(url)
	for i := 0; i < testCnt; i++ {
		tinyurl := createPutRequest(url + strconv.Itoa(i))
		if tinyurl == tinyurl_prev {
			panic("same tinyurl for different longurls")
		}
	}
	println("TestPut is passed")
}

func testGetGood() {
	url := "abacabadabacaba"
	tinyurl := createPutRequest(url)
	for i := 0; i < testCnt; i++ {
		response, err := http.Get(URL + tinyurl)

		if err != nil {
			panic("Could not do http GET request" + err.Error())
		}

		if response.StatusCode != 200 {
			panic("returned code not 200")
		}
	}
	println("TestGetGood is passed")
}

func testGetBad() {
	url := "random???"

	for i := 0; i < testCnt; i++ {
		response, err := http.Get(URL + url)
		if err != nil {
			panic("Could not do http GET request" + err.Error())
		}

		if response.StatusCode != 404 {
			panic("returned code not 404")
		}
	}
	println("TestGetBad is passed")
}

func main() {
	testPut()
	testGetGood()
	testGetBad()
}
