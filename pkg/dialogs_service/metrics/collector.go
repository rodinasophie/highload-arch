package metrics

import (
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

const MONITORING_ENABLED = true

var (
	latency  *prometheus.HistogramVec
	httpReqs *prometheus.CounterVec
)

func InitMetrics() {
	httpReqs = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "social_http_requests_total",
			Help: "How many HTTP requests processed, partitioned by status code and HTTP method.",
		},
		[]string{"code", "method", "name"},
	)

	latency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "social_request_latency_seconds",
			Help: "Request Latency in seconds.",
		},
		[]string{"method", "name"},
	)

	prometheus.MustRegister(httpReqs)
	prometheus.MustRegister(latency)
}

func IncRequests(errorCode int, method, name string) {
	if !MONITORING_ENABLED {
		return
	}
	httpReqs.WithLabelValues(strconv.Itoa(errorCode), method, name).Inc()
}

func AddLatencyValue(start time.Time, method, name string) {
	if !MONITORING_ENABLED {
		return
	}
	latency.WithLabelValues(method, name).Observe(time.Since(start).Seconds())
}
