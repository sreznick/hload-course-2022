package server

import (
	"main/internal/config"
	"main/internal/kafka"
	"main/internal/metrics"
	"main/internal/redis"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type server struct {
	router   *gin.Engine
	redis    redis.Interface
	producer kafka.Producer
}

func New(producer kafka.Producer, redis redis.Interface) *server {
	return &server{
		router:   gin.Default(),
		redis:    redis,
		producer: producer,
	}
}

func (s *server) Run() {
	s.router.Use(metrics.PrometheusMiddleware)
	s.router.GET("/:url", s.redirectUrl)
	s.router.GET("/metrics", func(c *gin.Context) {
		handler := promhttp.Handler()
		handler.ServeHTTP(c.Writer, c.Request)
	})

	s.router.Run(config.BaseURL)
}
