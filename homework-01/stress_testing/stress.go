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
	fmt.Println("Start create requests")
	urls := []string{
		"https://ya.ru",
		"https://google.com",
		"https://duckduckgo.com",
		"https://go.dev",
		"https://discord.com",
		"https://spbu.ru",
		"https://zoom.us",
		"https://python.org",
		"https://golangdocs.com",
		"https://gitlab.com",
	}

	for i := 0; i < createResponseCount; i++ {
		randomSuffix := RandomString(5, 6) // Approximate 77% will be unique
		longurl := Pick(urls) + randomSuffix
		requestCreate(longurl)
	}
	fmt.Println("Finish create requests")
}

func checkValidUrl(url string) {
	code := requestGet(url)
	if code != 302 {
		panic(fmt.Sprintf("Invalid response code %d, but it should be 302", code))
	}
}

func runGetSuccess() {
	fmt.Println("Start success get requests")
	testsForOneUrl := 100
	urlsCount := getSuccessResponseCount / testsForOneUrl
	for i := 0; i < urlsCount; i++ {
		longUrl := fmt.Sprintf("https://%s.ru", RandomString(20, 62))
		tinyUrl := requestCreate(longUrl)
		for j := 0; j < testsForOneUrl; j++ {
			checkValidUrl(tinyUrl)
		}
	}
	fmt.Println("Finish success get requests")
}

func checkInvalidUrl(url string) {
	code := requestGet(url)
	if code != 404 {
		panic(fmt.Sprintf("Code response %d, but it should be 404", code))
	}
}

func runGetInvalid() {
	fmt.Println("Start bad get requests")

	// 1. Invaid symbols in tiny url
	badTinyUrls := []string{
		"_______",
		"abdcef!",
		"аофыдвл",
		"abdcef",
		"a",
		"",
	}
	for _, badUrl := range badTinyUrls {
		checkInvalidUrl(badUrl)
	}

	// 2. Valid symbols but tinyUrl does not exists
	for i := 0; i < getSuccessResponseCount; i++ {
		badUrl := RandomString(6, 62) + "z" // I belive that we have less then (62^7 - 62^6) tiny urls

		checkInvalidUrl(badUrl)
	}
	fmt.Println("Finish bad get requests")
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
	fmt.Println("Finish")
}
