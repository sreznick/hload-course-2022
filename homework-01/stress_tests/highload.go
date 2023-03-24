package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
)

const (
	createURL               = "http://localhost:8080/create"
	getURL                  = "http://localhost:8080/%s"
	createResponseCount     = 10_000
	getSuccessResponseCount = 100_000
	getInvalidResponseCount = 100_000
)

type UrlSerializer struct {
	LongUrl string `json:"longurl"`
	TinyUrl string `json:"tinyurl"`
}

func requestCreate(longurl string) string {
	client := &http.Client{}

	var jsonData = []byte(fmt.Sprintf("{\"longurl\": \"%s\"}", longurl))
	request, _ := http.NewRequest("PUT", createURL, bytes.NewBuffer(jsonData))
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")

	response, err := client.Do(request)
	if err != nil {
		panic(err.Error())
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		panic(err.Error())
	}

	var responseJson UrlSerializer
	err1 := json.Unmarshal(body, &responseJson)
	if err1 != nil {
		panic(err1.Error())
	}

	return responseJson.TinyUrl
}

func requestGet(tinyurl string) int {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	request, err := http.NewRequest("GET", fmt.Sprintf(getURL, tinyurl), nil)
	if err != nil {
		panic(err)
	}
	response, err := client.Do(request)
	if err != nil {
		panic(err)
	}
	return response.StatusCode
}

func runCreate() {
	fmt.Println("Create requests testing")
	urls := []string{
		"https://digitalocean.com",
		"https://cloud.yandex.ru",
		"https://go.dev",
		"https://google.ru/",
		"https://spbu.ru",
		"https://isocpp.org/",
	}

	for i := 0; i < createResponseCount; i++ {
		randomSuffix := RandomString(5, 6)
		longurl := Pick(urls) + randomSuffix
		requestCreate(longurl)
	}
}

func checkValidUrl(url string) {
	code := requestGet(url)
	if code != 302 {
		panic(fmt.Sprintf("Invalid response code %d, but it should be 302", code))
	}
}

func runGetSuccess() {
	fmt.Println("Success get requests testing")
	testsForOneUrl := 100
	urlsCount := getSuccessResponseCount / testsForOneUrl
	for i := 0; i < urlsCount; i++ {
		longUrl := fmt.Sprintf("https://%s.ru", RandomString(20, 62))
		tinyUrl := requestCreate(longUrl)
		for j := 0; j < testsForOneUrl; j++ {
			checkValidUrl(tinyUrl)
		}
	}
}

func checkInvalidUrl(url string) {
	code := requestGet(url)
	if code != 404 {
		panic(fmt.Sprintf("Code response %d, but it should be 404", code))
	}
}

func runGetInvalid() {
	fmt.Println("Bad get requests testing")

	// tiny urls with invalid symbols
	badTinyUrls := []string{
		"____",
		"abdcef!",
		"abdcef",
		"a",
		"",
	}
	for _, badUrl := range badTinyUrls {
		checkInvalidUrl(badUrl)
	}

	// valid tiny urls which does not exist
	for i := 0; i < getInvalidResponseCount; i++ {
		badUrl := RandomString(6, 62) + "q"

		checkInvalidUrl(badUrl)
	}
}

func main() {
	fmt.Println("Start")
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		runCreate()
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		runGetSuccess()
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		runGetInvalid()
		wg.Done()
	}()

	wg.Wait()
	fmt.Println("Testing finished")
}
