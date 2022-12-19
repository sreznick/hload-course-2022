package all

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
        HTTPPutRequestsCount = promauto.NewCounter(prometheus.CounterOpts{
                Name: "http_put_requests_count",
                Help: "The total number of HTTP PUT requests",
        })
        HTTPGetRequestsCount = promauto.NewCounter(prometheus.CounterOpts{
                Name: "http_get_requests_count",
                Help: "The total number of HTTP GET requests",
        })
        HTTPPutRequestsDurations = prometheus.NewSummary(prometheus.SummaryOpts{
	            	Name: "http_put_requests_durations",
		            Help: "HTTP PUT requests durations",
	            	Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
	      })
        HTTPGetRequestsDurations = prometheus.NewSummary(prometheus.SummaryOpts{
	            	Name: "http_get_requests_durations",
		            Help: "HTTP GET requests durations",
	            	Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
        })
)
