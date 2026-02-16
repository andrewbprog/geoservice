package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

var (
	// HTTPRequestCount - кол-во HTTP-запросов
	HTTPRequestCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests per endpoint",
		},
		[]string{"method", "path"},
	)

	// HTTPRequestDuration - длительность HTTP-запросов
	HTTPRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests per endpoint",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	// CacheDuration - время доступа к кэшу
	CacheDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "cache_duration_seconds",
			Help:    "Cache access duration per method",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method"},
	)

	// DBDuration - время доступа к БД
	DBDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "db_duration_seconds",
			Help:    "DB access duration per method",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method"},
	)

	// APIDuration - время обращения к внешнему API
	APIDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "external_api_duration_seconds",
			Help:    "External API access duration per method",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method"},
	)
)

// Init — регистрируем метрики
func Init() {
	prometheus.MustRegister(
		HTTPRequestCount,
		HTTPRequestDuration,
		CacheDuration,
		DBDuration,
		APIDuration,
	)
}

// Handler для /metrics
func Handler() http.Handler {
	return promhttp.Handler()
}
