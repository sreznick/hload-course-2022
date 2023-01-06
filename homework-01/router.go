package main

import (
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

func createUrl(c *gin.Context) {
	longUrl, err := parseJSON(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"response": err})
		return
	}
	tinyUrl, err := AddUrl(longUrl)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"response": err})
		return
	}
	c.JSON(http.StatusOK, gin.H{"longurl": longUrl, "tinyurl": tinyUrl})
	return
}

func getTinyUrl(c *gin.Context, tinyUrl string) {
	longUrl, err := GetUrl(tinyUrl)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"response": err})
		return
	}

	c.Redirect(http.StatusFound, longUrl)
	return
}

func setupRouter() *gin.Engine {
	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	r.PUT("/create", func(c *gin.Context) {
		createOpsProcessed.Inc()
		startTime := time.Now()

		createUrl(c)

		createOpsTime.Observe(float64(time.Since(startTime).Microseconds()))
	})

	r.GET("/:tinyurl", func(c *gin.Context) {
		getOpsProcessed.Inc()
		startTime := time.Now()

		tinyUrl := c.Params.ByName("tinyurl")
		getTinyUrl(c, tinyUrl)

		getOpsTime.Observe(float64(time.Since(startTime).Microseconds()))
	})

	return r
}
