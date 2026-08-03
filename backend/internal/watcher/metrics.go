package watcher

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	watcherPollsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "watcher_polls_total",
		Help: "Total number of watch polls by result.",
	}, []string{"result"})

	watcherPollDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "watcher_poll_duration_seconds",
		Help:    "Watch poll duration in seconds.",
		Buckets: prometheus.DefBuckets,
	})

	watcherTracksRemovedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "watcher_tracks_removed_total",
		Help: "Total number of tracks detected as removed from source playlists.",
	})

	watcherTracksArchivedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "watcher_tracks_archived_total",
		Help: "Total number of tracks appended to archive playlists.",
	})

	watcherWatchesDue = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "watcher_watches_due",
		Help: "Number of due watches claimed in the latest tick.",
	})
)

func MetricsHandler() http.Handler {
	return promhttp.Handler()
}
