package obs

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics holds the prometheus collectors and the registry they're on.
type Metrics struct {
	registry     *prometheus.Registry
	requests     *prometheus.CounterVec
	requestTime  *prometheus.HistogramVec
	inFlightReqs prometheus.Gauge
}

// sets up the registry with the default Go/process collectors plus our HTTP ones
func NewMetrics() *Metrics {
	reg := prometheus.NewRegistry()
	reg.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)

	m := &Metrics{
		registry: reg,
		requests: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total HTTP requests by method, route, and status code.",
		}, []string{"method", "route", "status"}),
		requestTime: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds by method and route.",
			Buckets: prometheus.DefBuckets,
		}, []string{"method", "route"}),
		inFlightReqs: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "http_requests_in_flight",
			Help: "Number of HTTP requests currently being served.",
		}),
	}

	reg.MustRegister(m.requests, m.requestTime, m.inFlightReqs)
	return m
}

// Handler serves the metrics in prometheus text format.
func (m *Metrics) Handler() http.Handler {
	return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{})
}

// Middleware records a count and duration for every request. route is the fixed
// pattern the handler is mounted on (e.g. "/query"), passed in so high-cardinality
// paths don't blow up the label set.
func (m *Metrics) Middleware(route string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.inFlightReqs.Inc()
		defer m.inFlightReqs.Dec()

		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		start := time.Now()
		next.ServeHTTP(rec, r)
		elapsed := time.Since(start).Seconds()

		m.requests.WithLabelValues(r.Method, route, strconv.Itoa(rec.status)).Inc()
		m.requestTime.WithLabelValues(r.Method, route).Observe(elapsed)
	})
}
