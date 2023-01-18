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

func requestCreate(client *http.Client, longurl string) string {
	url := "http://localhost:8080/create"

	var jsonData = []byte(fmt.Sprintf(`{
		"longurl": "%s"
	}`, longurl))
	request, _ := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")

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

func requestGetGood(client *http.Client, tinyurl string) (int, string) {
	url := "http://localhost:8080/" + tinyurl

	resp, error := client.Get(url)

	if error != nil {
		panic(error)
	}
	red_url, error := resp.Location()
	if error != nil {
		panic(error)
	}

	return resp.StatusCode, red_url.Host
}

func requestGetBad(client *http.Client) int {
	url := "http://localhost:8080/" + "haskellcurry42"

	resp, error := client.Get(url)

	if error != nil {
		panic(error)
	}
	return resp.StatusCode
}

func main() {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	longurl := "https://codeforces.com/"
	short_url := requestCreate(client, longurl)
	fmt.Println(short_url)

	for i := 1; i <= 10000; i++ {
		requestCreate(client, longurl+fmt.Sprint(i))

		if i%100 == 0 {
			fmt.Printf("CREATE: %d/%d\n", i, 10000)
		}
	}

	for i := 1; i <= 100000; i++ {
		code, url := requestGetGood(client, short_url)
		if code != 302 {
			panic("returned code not 302")
		}
		if url != "codeforces.com" {
			panic("redirect location didn't match")
		}
		if i%1000 == 0 {
			fmt.Printf("GET GOOD: %d/%d\n", i, 100000)
		}
	}

	for i := 1; i <= 100000; i++ {
		code := requestGetBad(client)
		if code != 404 {
			panic("returned code not 404")
		}

		if i%1000 == 0 {
			fmt.Printf("GET BAD: %d/%d\n", i, 100000)
		}
	}
}
