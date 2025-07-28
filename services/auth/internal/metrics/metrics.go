package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	TotalRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "hcaas_http_requests_total",
			Help: "Total number of HTTP requests to the auth service",
		},
		[]string{"service", "method", "path", "status"},
	)

	RequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "hcaas_http_request_duration_seconds",
			Help:    "Histogram of response duration auth service",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"service", "method", "path", "status"},
	)
)

func Register() {
	prometheus.MustRegister(TotalRequests, RequestDuration)
}
