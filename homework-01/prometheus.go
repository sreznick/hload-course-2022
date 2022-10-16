package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	createOpsProcessed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "create_counter",
		Help: "The total number of processed events",
	})

	getOpsProcessed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "get_counter",
		Help: "The total number of processed events",
	})

	createOpsTime = promauto.NewSummary(prometheus.SummaryOpts{
		Name: "create_time",
		Help: "The time of processed events",
	})

	getOpsTime = promauto.NewSummary(prometheus.SummaryOpts{
		Name: "get_time",
		Help: "The time of processed events",
	})
)
