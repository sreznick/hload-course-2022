package main

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type LongUrlBody struct {
	LongUrl string `json:"longurl"`
}

func parseJSON(c *gin.Context) (string, error) {
	body := LongUrlBody{}
	err := c.BindJSON(&body)
	if err != nil {
		return "", err
	}
	return body.LongUrl, nil
}

func createUrl(c *gin.Context, db *sql.DB) {
	longUrl, err := parseJSON(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"response": err})
		return
	}
	tinyUrl, err := addUrl(db, longUrl)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"response": err})
		return
	}
	c.JSON(http.StatusOK, gin.H{"longurl": longUrl, "tinyurl": tinyUrl})
	return
}

func getTinyUrl(c *gin.Context, db *sql.DB, tinyUrl string) {
	longUrl, err := getUrl(db, tinyUrl)
	if err != nil {
		c.Writer.WriteHeader(404)
		return
	}

	c.Redirect(http.StatusFound, longUrl)
	return
}

func setupRouter(db *sql.DB) *gin.Engine {
	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	r.PUT("/create", func(c *gin.Context) {
		createOpsProcessed.Inc()
		startTime := time.Now()

		createUrl(c, db)

		createOpsTime.Observe(float64(time.Since(startTime).Microseconds()))
	})

	r.GET("/:tinyurl", func(c *gin.Context) {
		getOpsProcessed.Inc()
		startTime := time.Now()

		tinyUrl := c.Params.ByName("tinyurl")
		getTinyUrl(c, db, tinyUrl)

		getOpsTime.Observe(float64(time.Since(startTime).Microseconds()))
	})

	return r
}
