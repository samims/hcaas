package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	RequestCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "hcaas_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"path", "method", "status"},
	)

	RequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "hcaas_http_request_duration_seconds",
			Help:    "Histogram of response durations",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"path", "method"},
	)

	// URLCheckStatus background check stats
	URLCheckStatus = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "hcaas_url_check_status_total",
			Help: "Number of successful or failed URL checks",
		},
		[]string{"status"},
	)

	//URLCheckDuration checks how much time take to ping the url
	URLCheckDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "url_check_duration_seconds",
			Help:    "Duration of URL health checks",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"status"},
	)
)

func Init() {
	prometheus.MustRegister(RequestCount, RequestDuration, URLCheckStatus, URLCheckDuration)
}
