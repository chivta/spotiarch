package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	httpRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total number of HTTP requests.",
	}, []string{"method", "path", "status"})

	httpRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "HTTP request duration in seconds.",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "path"})

	httpRequestsInFlight = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "http_requests_in_flight",
		Help: "Current number of HTTP requests being served.",
	})
)

func MetricsMiddleware(skip ...string) fiber.Handler {
	skipMap := make(map[string]struct{}, len(skip))
	for _, path := range skip {
		skipMap[path] = struct{}{}
	}

	return func(c *fiber.Ctx) error {
		if _, ok := skipMap[c.Path()]; ok {
			return c.Next()
		}

		httpRequestsInFlight.Inc()
		defer httpRequestsInFlight.Dec()
		start := time.Now()
		err := c.Next()

		path := c.Route().Path
		if path == "" {
			path = c.Path()
		}
		httpRequestsTotal.WithLabelValues(c.Method(), path, strconv.Itoa(c.Response().StatusCode())).Inc()
		httpRequestDuration.WithLabelValues(c.Method(), path).Observe(time.Since(start).Seconds())
		return err
	}
}

func MetricsHandler() http.Handler {
	return promhttp.Handler()
}
