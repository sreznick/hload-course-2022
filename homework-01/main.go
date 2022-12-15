package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

const SQL_DRIVER = "postgres"
const (
	host     = "localhost"
	port     = 5432
	user     = "kirill"
	password = "111"
	dbname   = "kirill"
)

func setup() {
	db := setupModels()
	r := setupRouter(db)
	r.Run(":8081")
}

func setupModels() *sql.DB {
	psqlInfo := fmt.Sprintf("postgres://%v:%v@%v:%v/%v?sslmode=disable", user, password, host, port, dbname)
	conn, err := sql.Open(SQL_DRIVER, psqlInfo)
	if err != nil {
		fmt.Println("Failed to open", err)
		panic("exit")
	}

	err = conn.Ping()
	if err != nil {
		fmt.Println("Failed to ping database ", err)
		panic("exit")
	}
	_, err = conn.Exec("CREATE TABLE IF NOT EXISTS urls (id SERIAL, url TEXT PRIMARY KEY)")
	if err != nil {
		fmt.Println("Failed to create table", err)
		panic("exit")
	}
	return conn
}

func setupRouter(db *sql.DB) *gin.Engine {
	r := gin.Default()

	r.PUT("/create", func(c *gin.Context) {
		recordCreateStart()

		start := time.Now()

		handleCreate(c, db)

		recordCreateTime(float64(time.Since(start).Milliseconds()))
	})

	r.GET("/:url", func(c *gin.Context) {
		recordGetStart()

		start := time.Now()

		handleUrlGet(c, db)

		recordGetTime(float64(time.Since(start).Milliseconds()))
	})

	return r
}

func main() {
	stress()
}

type CreateResponse struct {
	Longurl string
	Tinyurl string
}

func stressCreate() string {
	jsonData := []byte(`{
		"longurl": "https://google.com"
	}`)

	request, err := http.NewRequest("PUT", "http://127.0.0.1:8081/create", bytes.NewBuffer(jsonData))

	if err != nil {
		panic(err)
	}

	request.Header.Set("Content-Type", "application/json; charset=UTF-8")

	client := &http.Client{}
	response, err := client.Do(request)

	if err != nil {
		panic(err)
	}

	defer response.Body.Close()

	body, _ := ioutil.ReadAll(response.Body)
	var responseJson CreateResponse
	json.Unmarshal(body, &responseJson)

	return responseJson.Tinyurl
}

func commitGood(url string) {
	resp, err := http.Get("http://127.0.0.1:8081/" + url)

	if err != nil {
		panic(err)
	}

	if resp.StatusCode != 200 {
		panic(fmt.Errorf("waited for ok code"))
	}
}

func commitBad() {
	resp, err := http.Get("http://127.0.0.1:8081/" + "try_a_bit_harder_next_time")
	if err != nil {
		panic(err)
	}

	if resp.StatusCode != 404 {
		panic(fmt.Errorf("waited for 404 code"))
	}
}

func stress() {
	basic := stressCreate()
	for i := 0; i <= 10000; i++ {
		newUrl := stressCreate()
		if newUrl != basic {
			panic(fmt.Errorf("url changed"))
		}
	}

	for i := 0; i <= 100000; i++ {
		commitGood(basic)
	}

	for i := 0; i <= 100000; i++ {
		commitBad()
	}
}
