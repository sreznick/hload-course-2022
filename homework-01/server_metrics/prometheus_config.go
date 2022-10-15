package server_metrics

import (
	_ "net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	_ "github.com/prometheus/client_golang/prometheus/promhttp"
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
		Help: "Time of `get` operation processing",
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
