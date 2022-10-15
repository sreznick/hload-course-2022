package main

import (
	"database/sql"
	"fmt"
	"net/http"
	_ "net/http"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	_ "github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	create   = "INSERT INTO urls (url) VALUES ($1) ON CONFLICT DO NOTHING"
	getByUrl = "SELECT id FROM urls WHERE url = $1"
	getById  = "SELECT url FROM urls WHERE id = $1"

	domainSize = 62
	spaceSize  = 62 ^ 7
)

type CreateRequestBody struct {
	Longurl string `json:"longurl"`
}

func handleCreate(context *gin.Context, db *sql.DB) {
	body := CreateRequestBody{}
	err := context.BindJSON(&body)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"response": "bad long url"})
		return
	}

	_, err = db.Exec(create, body.Longurl)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"response": "can't insert new long url: " + err.Error()})
		return
	}

	var id int
	err = db.QueryRow(getByUrl, body.Longurl).Scan(&id)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"response": "can't retrieve just inserted url id: " + err.Error()})
		return
	}
	shortUrl, err := idToUrl(id)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{"response": "all short urls are used"})
		return
	}

	context.JSON(http.StatusOK, gin.H{"lognurl": body.Longurl, "tinyurl": shortUrl})
}

func idToUrl(id int) (string, error) {
	if id >= spaceSize || id < 0 {
		return "", fmt.Errorf("id out of domain")
	}
	bytes := make([]byte, 7)
	for i, _ := range bytes {
		bytes[i] = fromIntToByte(id % domainSize)
		id /= domainSize
	}
	return string(bytes), nil
}

func fromIntToByte(id int) byte {
	if id < 26 {
		return byte('a' + id)
	} else if id >= 26 && id < 52 {
		return byte('A' + (id - 26))
	} else {
		return byte('0' + (id - 52))
	}
}

func handleUrlGet(context *gin.Context, db *sql.DB) {
	shortUrl := context.Params.ByName("url")
	id, err := fromBytesToInt([]byte(shortUrl))
	if err != nil {
		context.Writer.WriteHeader(404)
		return
	}

	var longUrl string
	err = db.QueryRow(getById, id).Scan(&longUrl)
	if err != nil {
		context.Writer.WriteHeader(404)
		return
	}

	context.Redirect(302, longUrl)
}

func fromBytesToInt(bytes []byte) (int, error) {
	if len(bytes) != 7 {
		return 0, fmt.Errorf("wrong bytes size")
	}
	result := 0
	for i := len(bytes) - 1; i >= 0; i-- {
		result *= domainSize
		id, err := fromByteToInt(bytes[i])
		if err != nil {
			return 0, err
		}
		result += id
	}
	return result, nil
}

func fromByteToInt(symbol byte) (int, error) {
	if symbol >= 'a' && symbol <= 'z' {
		return int(symbol - 'a'), nil
	}
	if symbol >= 'A' && symbol <= 'Z' {
		return int(symbol-'A') + 26, nil
	}
	if symbol >= '0' && symbol <= '9' {
		return int(symbol-'0') + 52, nil
	}
	return 0, fmt.Errorf("byte out of domain")
}

var (
	createQuery = promauto.NewCounter(prometheus.CounterOpts{
		Name: "create_counter",
		Help: "total number of `create` queries`",
	})
	createTime = promauto.NewSummary(prometheus.SummaryOpts{
		Name: "create_time",
		Help: "time of `create` query",
	})
	getQuery = promauto.NewCounter(prometheus.CounterOpts{
		Name: "get_counter",
		Help: "total number of `get` queries`",
	})
	getTime = promauto.NewSummary(prometheus.SummaryOpts{
		Name: "get_time",
		Help: "еime of `get` query",
	})
)

func recordGetStart() {
	getQuery.Inc()
}

func recordCreateStart() {
	createQuery.Inc()
}

func recordGetTime(time float64) {
	getTime.Observe(time)
}

func recordCreateTime(time float64) {
	createTime.Observe(time)
}
