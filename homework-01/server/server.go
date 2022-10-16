package server

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"main/database"
	"main/model"
	"net/http"
)

type PutRequestJson struct {
	Longurl string `json:"longurl"`
}

func getLongUrlFromJson(c *gin.Context) (string, error) {
	body := PutRequestJson{}
	err := c.BindJSON(&body)
	if err != nil {
		errorMessage := "bad json format"
		return "", fmt.Errorf(errorMessage)
	}

	return body.Longurl, nil
}

func create(c *gin.Context, db *database.UrlTable) {
	longUrl, err := getLongUrlFromJson(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"response": err.Error()})
		return
	}

	tinyUrl, modelError := model.GetTinyUrlByLongUrl(db, longUrl)
	if modelError != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"response": modelError.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"longurl": longUrl, "tinyurl": tinyUrl})
}

func doRedirectOnLongString(c *gin.Context, db *database.UrlTable, tinyUrl string) {
	longUrl, err := model.GetLongUrlByTinyUrl(db, tinyUrl)
	if err != nil {
		c.Writer.WriteHeader(http.StatusNotFound)
		return
	}

	c.Redirect(http.StatusFound, longUrl)
}

func SetupRouter(db *database.UrlTable) *gin.Engine {
	r := gin.Default()

	r.PUT("/create", func(c *gin.Context) {
		create(c, db)
	})

	r.GET("/:url", func(c *gin.Context) {
		tinyUrl := c.Params.ByName("url")
		doRedirectOnLongString(c, db, tinyUrl)
	})
	return r
}
