package spotify

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	APIDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "spotify_api_duration_seconds",
		Help:    "Spotify API call duration in seconds.",
		Buckets: prometheus.DefBuckets,
	}, []string{"method"})

	APIRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "spotify_api_requests_total",
		Help: "Total number of Spotify API requests by status.",
	}, []string{"method", "status"})
)
