package server_api

import (
	"fmt"
	"time"

	"net/http"

	_ "github.com/lib/pq"

	"github.com/gin-gonic/gin"

	dbi "main/db_interactor"

	gen "main/url_generator"

	met "main/server_metrics"
)

type CreateRequest struct {
	Longurl string `json:"longurl"`
}

type CreateResponse struct {
	Longurl  string `json:"longurl"`
	Shorturl string `json:"shorturl"`
}

func putCreate(c *gin.Context, db *dbi.DbInteractor) {
	var request CreateRequest

	if err := c.BindJSON(&request); err != nil {
		return
	}

	fmt.Println(request.Longurl)
	short_url := gen.GenerateShortUrl(request.Longurl)
	db.InsertURL(short_url, request.Longurl)
	var response = CreateResponse{request.Longurl, short_url}
	c.IndentedJSON(http.StatusCreated, response)
}

func getLongURL(c *gin.Context, db *dbi.DbInteractor) {
	shorturl := c.Param("shorturl")
	fmt.Println(shorturl)
	var longurl, err = db.GetLongURL(shorturl).Get()
	fmt.Println(longurl)
	if err != nil {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "url not found"})
	} else {
		c.Redirect(http.StatusFound, longurl)
	}
}

func SetupRouter(db *dbi.DbInteractor) *gin.Engine {
	router := gin.Default()
	router.GET(":shorturl", func(c *gin.Context) {
		met.RecordGetMetrics()

		start := time.Now()
		getLongURL(c, db)
		elapsed := float64(time.Since(start).Nanoseconds()) / 1000

		met.RecordGetTime(elapsed)
	})
	router.PUT("/create", func(c *gin.Context) {
		met.RecordCreateMetrics()

		start := time.Now()
		putCreate(c, db)

		elapsed := float64(time.Since(start).Nanoseconds()) / 1000

		met.RecordCreateTime(elapsed)

	})
	return router
}
