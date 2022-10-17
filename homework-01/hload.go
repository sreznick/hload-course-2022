package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"time"
	"urlShortener/utils"
	_ "urlShortener/utils"
)

func getAllShortURLs() []utils.UrlDB {
	b, err := ioutil.ReadFile("pass.conf")
	if err != nil {
		fmt.Print(err)
		res := []utils.UrlDB{}
		return res
	}

	// convert bytes to string
	pass := string(b)

	server := &utils.SSH{
		Ip:   "217.25.88.166",
		User: "root",
		Port: 22,
		Cert: pass,
	}

	err = server.Connect(utils.CERT_PASSWORD)
	if err != nil {
		fmt.Println(err)
		res := []utils.UrlDB{}
		return res
	}

	defer server.Close()

	utils.InitConection(*server)

	client := &utils.DBConnect{
		Ip:   "localhost",
		User: "postgres",
		Name: "url_shortener",
		Cert: pass}

	err = client.Open()

	if err != nil {
		fmt.Println(err)
		res := []utils.UrlDB{}
		return res
	}

	defer client.Close()

	utils.InitData(client)

	urls := client.GetURLS()

	return urls
}

func getLinks() []string {
	var res []string

	file, err := os.Open("links.txt")

	if err != nil {
		return res
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)
	// optionally, resize scanner's capacity for lines over 64K, see next example
	for scanner.Scan() {
		res = append(res, scanner.Text())
	}

	return res
}

func main() {

	links := getLinks()

	for i := 0; i < 10000; i++ {

		r := rand.Intn(100)
		time.Sleep(time.Duration(r) * time.Millisecond)

		link := links[rand.Intn(len(links))]
		url := utils.GetURL{URL: link}
		jsonResp, _ := json.Marshal(url)
		req, _ := http.NewRequest("PUT", "http://127.0.0.1:8000/create", bytes.NewBuffer(jsonResp))

		client := &http.Client{}
		_, err := client.Do(req)

		if err != nil {
			fmt.Println("Could not do put request", err)
			return
		}
	}

	shortURLs := getAllShortURLs()

	for i := 0; i < 100000; i++ {

		r := rand.Intn(100)
		time.Sleep(time.Duration(r) * time.Millisecond)

		shortKey := shortURLs[rand.Intn(len(shortURLs))].Key
		resp, err := http.Get("http://127.0.0.1:8000/" + shortKey)

		if resp != nil {
			resp.Body.Close()
		}

		if err != nil {
			fmt.Println("Could not do get request", err)
			return
		}
	}

	for i := 0; i < 100000; i++ {
		r := rand.Intn(100)
		time.Sleep(time.Duration(r) * time.Millisecond)

		_, err := http.Get("http://127.0.0.1:8000/" + utils.RandKey())
		if err != nil {
			fmt.Println("Could not do get request", err)
			return
		}
	}

}
