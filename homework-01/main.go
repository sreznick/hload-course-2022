package main

import (
    "database/sql"
    "fmt"
    "time"

    handlers "main/handlers"
    global "main/global"
    metrics "main/metrics"

    "github.com/gin-gonic/gin"
    _ "github.com/lib/pq"

    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

const SQL_DRIVER = "postgres"
const SQL_CONNECT_URL = "postgres://postgres:postgres@localhost"


func MetricsReporterMiddleware(c *gin.Context) {
  start := time.Now().UnixMicro()
  c.Next()
  end := time.Now().UnixMicro()
  total := float64(end - start)

  if c.Request.Method == "PUT" {
    metrics.HTTPPutRequestsCount.Inc()
    metrics.HTTPPutRequestsDurations.Observe(total)
  } else if c.Request.Method == "GET" {
    metrics.HTTPGetRequestsCount.Inc()
    metrics.HTTPGetRequestsDurations.Observe(total)
  }
}

func setupRouter() *gin.Engine {
	router := gin.New()

  router.Use(MetricsReporterMiddleware)

  router.PUT("/create", handlers.CreateTinyURL)
  router.GET("/:tinyurl", handlers.FetchLongURL)

	return router
}

func setupPrometheusRouter() *gin.Engine {
  router := gin.New()
  registry := prometheus.NewRegistry()
  registry.MustRegister(metrics.HTTPGetRequestsCount, metrics.HTTPPutRequestsCount, metrics.HTTPGetRequestsDurations, metrics.HTTPPutRequestsDurations)
  handler := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})

  router.GET("/metrics", gin.WrapH(handler))

  return router
}


func main() {
    connection, err := sql.Open(SQL_DRIVER, SQL_CONNECT_URL)
    if err != nil {
        fmt.Println("Failed to open", err)
        panic("exit")
    }
    global.Connection = connection

    err = connection.Ping()
    if err != nil {
        fmt.Println("Failed to ping database", err)
        panic("exit")
    }

	r := setupRouter()
  go r.Run(":8080")

  r = setupPrometheusRouter()
  r.Run(":2112")
}
