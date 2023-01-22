package server

import (
	"errors"
	"main/internal/config"
	"main/internal/models"
	"main/internal/postgres/consts"
	"main/internal/shortener"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (s *server) redirectUrl(c *gin.Context) {
	ctx := c.Request.Context()

	id, err := shortener.UrlToId(c.Param("url"))
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	longUrl, err := s.redis.GetUrl(ctx, id)
	if errors.Is(err, consts.ErrNotFound) {
		c.AbortWithStatus(http.StatusNotFound)
		return
	} else if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	clicks, err := s.redis.ChangeClicks(ctx, id, 1)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	if clicks > config.ClicksSend {
		_, err = s.redis.ChangeClicks(ctx, id, -config.ClicksSend)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		clicksKafka := models.Clicks{
			ShortUrlID: id,
			Inc:        config.ClicksSend,
		}

		err = s.producer.Produce(ctx, longUrl, clicksKafka)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
	}

	c.Redirect(http.StatusMovedPermanently, longUrl)
}
