package utils

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
)

var promRegisteredLinkCount = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name: "put_requests",
		Help: "Количество зарегистрированных ссылок",
	})

var requestProcessingTimeSummaryMs = prometheus.NewSummary(
	prometheus.SummaryOpts{
		Name:       "request_processing_time_summary_ms",
		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
	})

var requestProcessingTimeHistogramMs = prometheus.NewHistogram(
	prometheus.HistogramOpts{
		Name:    "request_processing_time_histogram_ms",
		Buckets: prometheus.LinearBuckets(0, 10, 20),
	})

var promReceivedLinkCount = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name: "get_requests",
		Help: "Количество запросов ссылок",
	})

func RegPrometheus() {
	prometheus.MustRegister(promReceivedLinkCount)
	prometheus.MustRegister(promRegisteredLinkCount)
	prometheus.MustRegister(requestProcessingTimeSummaryMs)
	prometheus.MustRegister(requestProcessingTimeHistogramMs)
}

func PrometheusPush() {

	if err := push.New("http://217.25.88.166:9091", "register_link").
		Collector(promRegisteredLinkCount).
		Grouping("urls", "create").
		Push(); err != nil {
		fmt.Println("Could not push completion time to Pushgateway:", err)
	}

	if err := push.New("http://217.25.88.166:9091", "receive_link").
		Collector(promReceivedLinkCount).
		Grouping("urls", "get").
		Push(); err != nil {
		fmt.Println("Could not push completion time to Pushgateway:", err)
	}
}
