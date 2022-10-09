package server_backend

import (
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	"main/url_backend"
	"net/http"
	"time"
)

const (
	insertNewUrl = "insert into urls(url) values ($1) on conflict do nothing"
	selectByUrl  = "select id from urls where url = $1"
)

type CreateRequestJsonBody struct {
	Longurl string `json:"longurl"`
}

func getLongUrlFromJson(c *gin.Context) (string, error) {
	body := CreateRequestJsonBody{}
	err := c.BindJSON(&body)
	if err != nil {
		errorMessage := "wrong json format"
		return "", fmt.Errorf(errorMessage)
	}

	return body.Longurl, nil
}

func create(c *gin.Context, db *sql.DB) {
	longUrl, err := getLongUrlFromJson(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"response": err})
		return
	}

	_, err = db.Exec(insertNewUrl, longUrl)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"response": "something wrong with database: " + err.Error()})
		return
	}

	var tinyUrlId int
	err = db.QueryRow(selectByUrl, longUrl).Scan(&tinyUrlId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"response": "something wrong with database: " + err.Error()})
		return
	}
	tinyUrl, err := url_backend.IdToUrl(tinyUrlId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"response": "Impossible to return tiny url: no more space for unique urls"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"longurl": longUrl, "tinyurl": tinyUrl})
}

func getUrl(c *gin.Context, db *sql.DB, urlVarName string) {
	shortUrl := c.Params.ByName(urlVarName)
	shortUrlId, err := url_backend.UrlToId(shortUrl)
	if err != nil {
		c.Writer.WriteHeader(404)
		return
	}
	var longUrl string

	err = db.QueryRow("select url from urls where id = $1", shortUrlId).Scan(&longUrl)
	if err != nil {
		c.Writer.WriteHeader(404)
		return
	}

	c.Redirect(302, longUrl)
}

func SetupRouter(db *sql.DB) *gin.Engine {
	r := gin.Default()

	r.PUT("/create", func(c *gin.Context) {
		RecordCreateMetrics()

		start := time.Now()
		create(c, db)
		elapsed := float64(time.Since(start).Nanoseconds()) / 1000

		RecordCreateTime(elapsed)
	})

	urlVarName := "url"
	r.GET(fmt.Sprintf("/:%s", urlVarName), func(c *gin.Context) {
		RecordGetMetrics()

		start := time.Now()
		getUrl(c, db, urlVarName)
		elapsed := float64(time.Since(start).Nanoseconds()) / 1000

		RecordGetTime(elapsed)
	})

	return r
}
