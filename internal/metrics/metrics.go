package metrics

import (
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	menuRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "coffee",
			Name:      "menu_requests_total",
			Help:      "Total number of requests to /v1/menu.",
		},
		[]string{"status"},
	)

	menuLatency = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: "coffee",
			Name:      "menu_latency_seconds",
			Help:      "Latency of /v1/menu requests in seconds.",
			Buckets:   append([]float64{0.002, 0.003, 0.004}, prometheus.DefBuckets...),
		},
	)
)

func init() {
	prometheus.MustRegister(menuRequests, menuLatency)
}

// ObserveMenu records a single /menu call with final status and duration.
func ObserveMenu(statusCode int, d time.Duration) {
	menuRequests.WithLabelValues(strconv.Itoa(statusCode)).Inc()
	menuLatency.Observe(d.Seconds())
}
