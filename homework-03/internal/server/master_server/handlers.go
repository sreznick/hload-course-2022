package server

import (
	"errors"
	"fmt"
	"main/internal/config"
	"main/internal/models"
	"main/internal/postgres/consts"
	"main/internal/shortener"
	"net/http"

	"github.com/gin-gonic/gin"
)

type createRequest struct {
	LongUrl string `json:"longurl"`
}

func (s *server) createTinyUrl(c *gin.Context) {
	ctx := c.Request.Context()

	var request createRequest
	c.BindJSON(&request)

	id, err := s.postgres.UpsertUrl(request.LongUrl)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	create := models.Create{
		LongUrl:    request.LongUrl,
		ShortUrlID: id,
	}

	if err = s.producer.Produce(ctx, fmt.Sprintf("%v", id), create); err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	tinyUrl := shortener.IdToUrl(id)
	c.JSON(http.StatusOK, gin.H{
		"tinyurl": config.BaseURL + "/" + tinyUrl,
		"longurl": request.LongUrl,
	})
}

func (s *server) redirectUrl(c *gin.Context) {
	ctx := c.Request.Context()
	id, err := shortener.UrlToId(c.Param("url"))
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	longUrl, err := s.postgres.GetUrl(id)
	if errors.Is(err, consts.ErrNotFound) {
		c.AbortWithStatus(http.StatusNotFound)
		return
	} else if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	if err = s.postgres.IncClicks(ctx, id, 1); err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.Redirect(http.StatusMovedPermanently, longUrl)
}
