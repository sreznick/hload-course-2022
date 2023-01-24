package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

var urls = []string{
	"https://codeforces.com/",
	"https://vk.com/",
	"https://ya.ru/",
	"https://huawei.com/",
	"https://arm.com/",
	"https://intel.com/",
	"https://nvidia.com/",
	"https://amd.com/",
	"https://youtube.com/",
	"https://google.com/",
	"https://mk.ru/",
	"https://ficbook.net/",
	"https://sberbank.ru/",
	"https://drom.ru/",
	"https://championat.com/",
	"https://kommersant.ru/",
	"https://hh.ru/",
	"https://music.yandex/",
	"https://drive2.ru/",
	"https://sports.ru/",
	"https://samsung.com/",
	"https://mos.ru/",
	"https://tiktok.com/",
}

var good = []string{}
var bad = []string{}
var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func initRnd() {
	rand.Seed(time.Now().UnixNano())
}

func getRandomUrl() string {
	id := rand.Int() % len(urls)
	return urls[id]
}

func getRandomBadUrl() string {
	id := rand.Int() % len(bad)
	return bad[id]
}

func getRandomGoodUrl() string {
	id := rand.Int() % len(good)
	return good[id]
}

func randString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func generateRandomBadUrls() error {
	for {
		url := "http://localhost:8080/" + randString(7)

		resp, err := http.Get(url)

		if err != nil {
			return err
		}
		if resp.StatusCode == 404 {
			bad = append(bad, url)
		}

		if len(bad) == 150 {
			break
		}
	}

	return nil
}

// Tests

type CreateResponse struct {
	Longurl string `json:"longurl"`
	Tinyurl string `json:"tinyurl"`
}

func requestCreate() string {
	url := "http://localhost:8080/create"
	addedUrl := getRandomUrl()

	var jsonData = []byte(fmt.Sprintf(`{
		"longurl": "%s"
	}`, addedUrl))
	request, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(jsonData))
	if err != nil {
		panic(err)
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		panic(err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			panic("Error while closing the body: " + err.Error())
		}
	}(response.Body)

	var responseJson CreateResponse
	err = json.NewDecoder(response.Body).Decode(&responseJson)
	if err != nil {
		return ""
	}

	if responseJson.Tinyurl == "" {
		panic("aaaaa " + responseJson.Longurl)
	}

	return responseJson.Tinyurl
}

func requestGetGood() int {
	url := "http://localhost:8080/" + getRandomGoodUrl()
	c := http.Client{CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	},
	}
	resp, err := c.Get(url)

	if err != nil {
		panic(err)
	}
	return resp.StatusCode
}

func requestGetBad() int {
	url := getRandomBadUrl()

	resp, err := http.Get(url)

	if err != nil {
		panic(err)
	}
	return resp.StatusCode
}

func main() {
	initRnd()
	err := generateRandomBadUrls()
	if err != nil {
		panic(err)
	}
	var wg sync.WaitGroup

	//Here database is empty
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 1000; i++ {
				code := requestGetBad()
				if code != 404 {
					fmt.Printf("returned code not 404 %d\n", i)
				}
			}
		}()
	}
	wg.Wait()
	fmt.Println("Done with bad urls")

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 10000; i++ {
				_ = requestCreate()
			}
		}()
	}
	wg.Wait()
	fmt.Println("Done with create")

	// build good array
	for i := 0; i < 20; i++ {
		url := requestCreate()
		good = append(good, url)
	}

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 1000; i++ {
				code := requestGetGood()
				if code != 302 {
					fmt.Printf("returned code not 302 %d\n", code)
				}
			}
		}()
	}
	wg.Wait()
	fmt.Println("Done with good urls")
}
