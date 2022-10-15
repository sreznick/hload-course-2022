package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type CreateResponse struct {
	Longurl  string `json:"longurl"`
	Shorturl string `json:"shorturl"`
}

func requestCreate() string {
	url := "http://localhost:8080/create"

	var jsonData = []byte(`{
		"longurl": "https://codeforces.com/"
	}`)
	request, _ := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")

	client := &http.Client{}
	response, error := client.Do(request)
	if error != nil {
		panic(error)
	}
	defer response.Body.Close()

	body, _ := ioutil.ReadAll(response.Body)
	var responseJson CreateResponse
	json.Unmarshal(body, &responseJson)

	return responseJson.Shorturl
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
	url := "http://localhost:8080/" + "##$@#$"

	resp, error := http.Get(url)

	if error != nil {
		panic(error)
	}
	return resp.StatusCode
}

func main() {
	short_url := requestCreate()
	fmt.Println(short_url)

	for i := 1; i <= 10000; i++ {
		url := requestCreate()
		if url != short_url {
			panic("new url appeared")
		}

		if i%100 == 0 {
			fmt.Printf("CREATE: %d/%d\n", i, 10000)
		}
	}

	for i := 1; i <= 100000; i++ {
		code := requestGetGood(short_url)
		if code != 200 {
			panic("returned code not 200")
		}

		if i%1000 == 0 {
			fmt.Printf("GET GOOD: %d/%d\n", i, 100000)
		}
	}

	for i := 1; i <= 100000; i++ {
		code := requestGetBad()
		if code != 404 {
			panic("returned code not 404")
		}

		if i%1000 == 0 {
			fmt.Printf("GET BAD: %d/%d\n", i, 100000)
		}
	}
}
