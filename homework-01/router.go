package main

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	putRequestCnt = promauto.NewCounter(prometheus.CounterOpts{
		Name: "put_request_cnt",
		Help: "Number of /create PUT requests",
	})
	getRequestCnt = promauto.NewCounter(prometheus.CounterOpts{
		Name: "get_request_cnt",
		Help: "Number of /:tinyurl GET requests",
	})
	putRequestTime = promauto.NewSummary(prometheus.SummaryOpts{
		Name: "put_request_time",
		Help: "Time of /create PUT request",
	})

	getRequestTime = promauto.NewSummary(prometheus.SummaryOpts{
		Name: "get_request_time",
		Help: "Time of /:tinyurl GET request",
	})
)

const (
	findByUrl = "SELECT id FROM urlsStorage WHERE long_url = $1"
	findById  = "SELECT long_url FROM urlsStorage WHERE id = $1"
	insertUrl = "INSERT INTO urlsStorage(long_url) VALUES ($1) ON CONFLICT DO NOTHING"
)

type PutRequestBody struct {
	Longurl string `json:"longurl"`
}

func putRequest(db *sql.DB, c *gin.Context) {
	request := PutRequestBody{}
	err := c.BindJSON(&request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"response": "wrong longurl"})
		return
	}

	longUrl := request.Longurl
	var id int

	err = db.QueryRow(findByUrl, longUrl).Scan(&id)
	if err != nil {
		_, insert_err := db.Exec(insertUrl, longUrl)
		if insert_err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"response": "insert error " + err.Error()})
			return
		}
		_ = db.QueryRow(findByUrl, longUrl).Scan(&id)
	}

	tinyUrl, err := IdToTinyUrl(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"response": "wrong id"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"longurl": longUrl, "tinyurl": tinyUrl})
}

func getRequest(db *sql.DB, c *gin.Context) {
	tinyUrl := c.Params.ByName("tinyurl")
	id, err := TinyUrlToId(tinyUrl)
	if err != nil {
		c.Writer.WriteHeader(http.StatusNotFound)
		return
	}

	var longUrl string
	err = db.QueryRow(findById, id).Scan(&longUrl)
	if err != nil {
		c.Writer.WriteHeader(http.StatusNotFound)
		return
	}

	c.Redirect(http.StatusFound, longUrl)
}

func SetupRouter(db *sql.DB) *gin.Engine {
	r := gin.Default()

	r.PUT("/create", func(c *gin.Context) {
		putRequestCnt.Inc()
		start := time.Now()
		putRequest(db, c)
		putRequestTime.Observe(float64(time.Since(start).Milliseconds()))
	})

	r.GET("/:tinyurl", func(c *gin.Context) {
		getRequestCnt.Inc()
		start := time.Now()
		getRequest(db, c)
		getRequestTime.Observe(float64(time.Since(start).Milliseconds()))
	})

	return r
}
