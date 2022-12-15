package server_backend

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	_ "github.com/prometheus/client_golang/prometheus/promhttp"
	_ "net/http"
)

func RecordGetMetrics() {
	getOpProcessed.Inc()
}

func RecordCreateMetrics() {
	createOpProcessed.Inc()
}

func RecordGetTime(time float64) {
	getOpTime.Observe(time)
}

func RecordCreateTime(time float64) {
	createOpTime.Observe(time)
}

var (
	createOpProcessed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "create_op_counter",
		Help: "The total number of processed `create` queries`",
	})

	createOpTime = promauto.NewSummary(prometheus.SummaryOpts{
		Name: "create_op_time",
		Help: "Time of `create` operation processing",
	})

	getOpProcessed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "get_op_counter",
		Help: "The total number of processed `get` queries`",
	})

	getOpTime = promauto.NewSummary(prometheus.SummaryOpts{
		Name: "get_op_time",
		Help: "Time of `get` operation processing",
	})
)

func SetupPrometheusRouter() *gin.Engine {
	r := gin.Default()
	registry := prometheus.NewRegistry()
	registry.MustRegister(createOpProcessed, createOpTime, getOpProcessed, getOpTime)
	handler := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})

	r.GET("/metrics", gin.WrapH(handler))
	return r
}
