package server

import (
	"main/internal/config"
	"main/internal/kafka"
	"main/internal/metrics"
	"main/internal/postgres"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type server struct {
	router   *gin.Engine
	postgres postgres.Interface
	producer kafka.Producer
}

func New(producer kafka.Producer, postgres postgres.Interface) *server {
	return &server{
		router:   gin.Default(),
		postgres: postgres,
		producer: producer,
	}
}

func (s *server) Run() {
	s.router.Use(metrics.PrometheusMiddleware)

	s.router.PUT("/create", s.createTinyUrl)
	s.router.GET("/:url", s.redirectUrl)
	s.router.GET("/metrics", func(c *gin.Context) {
		handler := promhttp.Handler()
		handler.ServeHTTP(c.Writer, c.Request)
	})

	s.router.Run(config.BaseURL)
}
