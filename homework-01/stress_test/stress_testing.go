package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

type CreateResponse struct {
	Longurl string `json:"longurl"`
	Tinyurl string `json:"tinyurl"`
}

func requestCreate() string {
	url := "http://localhost:8080/create"

	var jsonData = []byte(`{
		"longurl": "https://codeforces.com/"
	}`)
	request, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(jsonData))
	if err != nil {
		panic(err)
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		panic(err)
	}
	defer response.Body.Close()

	var responseJson CreateResponse
	err = json.NewDecoder(response.Body).Decode(&responseJson)
	if err != nil {
		panic(err)
	}

	if responseJson.Tinyurl == "" {
		panic("aaaaa " + responseJson.Longurl)
	}

	return responseJson.Tinyurl
}

func requestGetGood(tinyurl string) int {
	url := "http://localhost:8080/" + tinyurl

	resp, error := http.Get(url)

	if error != nil {
		panic(error)
	}
	return resp.StatusCode
}

func requestGetBad() int {
	url := "http://localhost:8080/" + "aboba@@&_"

	resp, error := http.Get(url)

	if error != nil {
		panic(error)
	}
	return resp.StatusCode
}

func main() {
	tinyUrl := requestCreate()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 10000; i++ {
			url := requestCreate()
			if url != tinyUrl {
				panic(fmt.Sprintf("new url appeared: %s %s %d", url, tinyUrl, i))
			}
		}
	}()

	wg.Add(100)
	for i := 0; i < 100; i++ {
		go func() {
			defer wg.Done()
			for i := 0; i < 1000; i++ {
				code := requestGetGood(tinyUrl)
				if code != 200 {
					panic(fmt.Sprintf("returned code not 200 %d", i))
				}
			}
		}()
	}

	wg.Add(100)
	for i := 0; i < 100; i++ {
		go func() {
			defer wg.Done()
			for i := 0; i < 1000; i++ {
				code := requestGetBad()
				if code != 404 {
					panic(fmt.Sprintf("returned code not 404 %d", i))
				}
			}
		}()
	}

	wg.Wait()
}
