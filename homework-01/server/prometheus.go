package server

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

func madeCreateOperationWithTimeInMCS(time float64) {
	createTimeSummary.Observe(time)
}

func madeGetOperationWithTimeInMCS(time float64) {
	getTimeSummary.Observe(time)
}

var (
	createTimeSummary = promauto.NewSummary(prometheus.SummaryOpts{
		Name: "create_processed_op_time_mcs",
		Help: "Duration of /create request",
	})

	getTimeSummary = promauto.NewSummary(prometheus.SummaryOpts{
		Name: "url_processed_op_time_mcs",
		Help: "Duration of GET /<tinyurl> request",
	})
)
