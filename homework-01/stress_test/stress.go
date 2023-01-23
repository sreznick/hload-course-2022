package main

import (
	"bytes"
	"fmt"
	"main/model"
	"net/http"
	"strings"
)

func makeCreateRequest(client *http.Client, i int) {
	bodyJSON := []byte(fmt.Sprintf(`{"longurl": "http://yandex.ru/%d"}`, i))
	request, err := http.NewRequest("PUT", "http://localhost:8080/create", bytes.NewBuffer(bodyJSON))
	if err != nil {
		panic(err)
	}
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")

	res, err := client.Do(request)

	if err != nil {
		panic(err)
	}

	if res.StatusCode != 200 {
		panic(res.StatusCode)
	}
}

func makeSuccessfulGetUrlRequest(client *http.Client, yandexId int) *http.Response {
	url, err := model.TinyUrlIdToTinyUrl(int64(yandexId))
	if err != nil {
		panic(err)
	}

	address := fmt.Sprintf("http://localhost:8080/%s", url)
	res, err := client.Get(address)

	if err != nil {
		panic(err)
	}

	if res.StatusCode != 302 {
		panic(res.StatusCode)
	}
	return res
}

func makeUnsuccessfulGetUrlRequest(number int) {
	res, err := http.Get(
		fmt.Sprintf(`http://localhost:8080/weirdoo/%d`, number))

	if err != nil {
		panic(err)
	}

	if res.StatusCode != 404 {
		panic(res.StatusCode)
	}
}

func testCreate(client *http.Client) {
	const createCount = 10000
	for i := 1; i <= createCount; i++ {
		makeCreateRequest(client, i)
		if i%1000 == 0 {
			fmt.Println("Create ", i)
		}
	}
}

func testSuccessfulGet(client *http.Client) {
	const successfulGetCount = 1
	for i := 1; i <= 50; i++ {
		makeCreateRequest(client, i)
	}
	for i := 1; i <= successfulGetCount; i++ {
		res := makeSuccessfulGetUrlRequest(client, i%49+1)
		location, err := res.Location()
		if err != nil {
			panic(err)
		}
		if !strings.HasPrefix(location.String(), "http://yandex.ru") {
			panic(location.String())
		}
		if i%100 == 0 {
			fmt.Println("Get successful", i)
		}
	}
}

func testUnsuccessfulGet() {
	const unsuccessfulGetCount = 100000
	for i := 1; i <= unsuccessfulGetCount; i++ {
		makeUnsuccessfulGetUrlRequest(i)
		if i%1000 == 0 {
			fmt.Println("Get unsuccessful", i)
		}
	}
}

func main() {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}}
	testCreate(client)
	testSuccessfulGet(client)
	testUnsuccessfulGet()
}
