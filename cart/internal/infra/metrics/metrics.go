// Package metrics ...
package metrics

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Количество запросов
	requestCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "cart",
		Name:      "handler_request_total",
		Help:      "Total count of request",
	}, []string{"handler", "type"})

	// Время исполнения запросов
	requestDurationHistogram = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "cart",
		Name:      "handler_request_duration_seconds",
		Help:      "Total duration of handler processing",
		Buckets:   prometheus.DefBuckets,
	}, []string{"handler", "status", "type"})

	// Количество элементов repository
	repoSizeGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "cart",
		Name:      "repo_size_total",
		Help:      "Size of repo",
	})
)

// IncRequestCount ...
func IncRequestCount(handler string, typeRequest string) {
	requestCounter.WithLabelValues(handler, typeRequest).Inc()
}

// RequestDuration ...
func RequestDuration(handler string, statusCode any, typeRequest string, duration time.Duration) {
	codeInt, ok := statusCode.(int)
	if ok {
		requestDurationHistogram.WithLabelValues(handler, http.StatusText(codeInt), typeRequest).Observe(float64(duration.Seconds()))
	}

	codeString, ok := statusCode.(string)
	if ok {
		requestDurationHistogram.WithLabelValues(handler, codeString, typeRequest).Observe(float64(duration.Seconds()))
	}
	//requestDurationHistogram.WithLabelValues(handler, http.StatusText(statusCode), typeRequest).Observe(float64(duration.Seconds()))
}

// StoreRepoSize ...
func StoreRepoSize(size float64) {
	repoSizeGauge.Set(size)
}
