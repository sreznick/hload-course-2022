package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
)

type PutURL struct {
	LongURL string `json:"longurl" binding:"required"`
	// short string `json:"shorturl" binding:"required"`
}

type ResponseForm struct {
	LongURL string `json:"longurl"`
	ShortURL string `json:"tinyurl"`
}

const (
	baseLongURL = "https://youtu.be/dQw4w9WgXcQ"
	invalidURL = "this is invalid"
)

func sendPut(longURL string) string {
	req_body := PutURL{ longURL }
	data, err := json.Marshal(req_body)
	if err != nil {
		panic(err.Error())
	}

	client := &http.Client{}
	request, err := http.NewRequest(http.MethodPut, "http://localhost:8080/create", bytes.NewBuffer(data))
	if err != nil {
		panic(err.Error())
	}
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")

	response, err := client.Do(request)
	if err != nil {
		panic(err.Error())
	}
	defer response.Body.Close()

	res_body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		panic(err.Error())
	}

	var responseForm ResponseForm
	err = json.Unmarshal(res_body, &responseForm)
	if err != nil {
		panic(err.Error())
	}

	return responseForm.ShortURL
}

func sendGet(shortURL string) int {
	response, err := http.Get("http://localhost:8080/" + shortURL)
	if response != nil {
		response.Body.Close()
	}
	if err != nil {
		panic(err.Error())
	}
	return response.StatusCode
}

func main() {
	shortURL := sendPut(baseLongURL)
	var wg sync.WaitGroup


	wg.Add(1)
	go func() {
		defer wg.Done()
		for k := 1; k <= 10000; k++ {
		// for k := 1; k <= 1; k++ {
			sendPut(baseLongURL)
			if k % 1000 == 0 {
				fmt.Println("%i create requests", k)
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for k := 1; k <= 100000; k++ {
		// for k := 1; k <= 1; k++ {
			statCode := sendGet(shortURL)
			if statCode >= 400 {
				fmt.Println("!statCode is %i", statCode)
			}
			if k % 1000 == 0 {
				fmt.Println("%i get requests", k)
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for k := 1; k <= 100000; k++ {
		// for k := 1; k <= 1; k++ {
			statCode := sendGet(invalidURL)
			if statCode < 404 {
				fmt.Println("!statCode is %i", statCode)
			}
			if k % 1000 == 0 {
				fmt.Println("%i invalid get requests", k)
			}
		}
	}()

	wg.Wait()
}
