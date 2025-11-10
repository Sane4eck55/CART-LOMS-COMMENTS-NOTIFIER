// Package metrics ...
package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Количество запросов
	requestCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "loms",
		Name:      "handler_request_total",
		Help:      "Total count of request",
	}, []string{"handler", "type"})

	// Время исполнения запросов
	requestDurationHistogram = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "loms",
		Name:      "handler_request_duration_seconds",
		Help:      "Total duration of handler processing",
		Buckets:   prometheus.DefBuckets,
	}, []string{"handler", "status", "type"})
)

// IncRequestCount ...
func IncRequestCount(handler string, typeRequest string) {
	requestCounter.WithLabelValues(handler, typeRequest).Inc()
}

// RequestDuration ...
func RequestDuration(handler string, statusCode string, typeRequest string, duration time.Duration) {
	requestDurationHistogram.WithLabelValues(handler, statusCode, typeRequest).Observe(float64(duration.Seconds()))
}
