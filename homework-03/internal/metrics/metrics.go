package metrics

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	totalRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "service_requests_total",
			Help: "Number of requests.",
		},
		[]string{"path"},
	)

	responseStatus = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "service_requests_status",
			Help: "Status of HTTP response",
		},
		[]string{"status"},
	)

	httpDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "service_requests_time_seconds",
			Help: "Duration of HTTP requests.",
		},
		[]string{"path"},
	)
)

func init() {
	prometheus.Register(totalRequests)
	prometheus.Register(responseStatus)
	prometheus.Register(httpDuration)
}

func PrometheusMiddleware(c *gin.Context) {
	timer := prometheus.NewTimer(httpDuration.WithLabelValues(c.Request.URL.Path))
	totalRequests.WithLabelValues(c.Request.URL.Path).Inc()

	c.Next()
	responseStatus.WithLabelValues(strconv.Itoa(c.Writer.Status())).Inc()

	timer.ObserveDuration()
}
