package server

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	putRequestsNumber = promauto.NewCounter(prometheus.CounterOpts{
		Name: "total_put_requests",
		Help: "The total number of processed /create PUT requests",
	})
	getRequestsNumber = promauto.NewCounter(prometheus.CounterOpts{
		Name: "total_get_requests",
		Help: "The total number of processed /:tinyurl GET requests",
	})
	putRequestTime = promauto.NewSummary(prometheus.SummaryOpts{
		Name: "put_request_time",
		Help: "Time of  /create PUT request processing",
	})

	getRequestTime = promauto.NewSummary(prometheus.SummaryOpts{
		Name: "get_request_time",
		Help: "Time of /:tinyurl GET request processing",
	})
)
