package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics holds all the Prometheus metrics for our service
type Metrics struct {
	// Request counters
	HTTPRequestsTotal    *prometheus.CounterVec
	HTTPRequestDuration  *prometheus.HistogramVec
	HTTPRequestsInFlight *prometheus.GaugeVec

	// Business logic metrics
	CampaignsDelivered *prometheus.CounterVec
	DatabaseQueries    *prometheus.CounterVec
	DatabaseErrors     *prometheus.CounterVec

	// Health check metrics
	HealthCheckStatus *prometheus.GaugeVec
}

// CachedMetrics wraps Metrics with pre-cached common metric combinations
type CachedMetrics struct {
	*Metrics

	// Pre-cached HTTP request metrics for common endpoints
	// Delivery endpoint metrics
	deliveryRequests200 prometheus.Counter
	deliveryRequests400 prometheus.Counter
	deliveryRequests500 prometheus.Counter
	deliveryDuration    prometheus.Observer
	deliveryInFlight    prometheus.Gauge

	// Health endpoint metrics
	healthRequests200 prometheus.Counter
	healthRequests500 prometheus.Counter
	healthDuration    prometheus.Observer
	healthInFlight    prometheus.Gauge

	// Pre-cached database metrics
	dbCampaignsSelect      prometheus.Counter
	dbTargetingRulesSelect prometheus.Counter
	dbQueryError           prometheus.Counter

	// Pre-cached health check metrics
	healthCheckDB    prometheus.Gauge
	healthCheckCache prometheus.Gauge
}

// NewPrometheusMetrics creates and registers all Prometheus metrics
func NewPrometheusMetrics() *Metrics {
	metrics := &Metrics{
		// HTTP request metrics
		HTTPRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "adbeacon_http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "endpoint", "status_code"},
		),

		HTTPRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "adbeacon_http_request_duration_seconds",
				Help:    "HTTP request duration in seconds",
				Buckets: prometheus.DefBuckets, // Standard buckets: 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10
			},
			[]string{"method", "endpoint"},
		),

		HTTPRequestsInFlight: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "adbeacon_http_requests_in_flight",
				Help: "Current number of HTTP requests being processed",
			},
			[]string{"method", "endpoint"},
		),

		// Business metrics
		CampaignsDelivered: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "adbeacon_campaigns_delivered_total",
				Help: "Total number of campaigns delivered",
			},
			[]string{"app", "country", "os"},
		),

		DatabaseQueries: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "adbeacon_database_queries_total",
				Help: "Total number of database queries",
			},
			[]string{"operation", "table"},
		),

		DatabaseErrors: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "adbeacon_database_errors_total",
				Help: "Total number of database errors",
			},
			[]string{"operation", "error_type"},
		),

		// Health check metrics
		HealthCheckStatus: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "adbeacon_health_check_status",
				Help: "Health check status (1 = healthy, 0 = unhealthy)",
			},
			[]string{"check_type"},
		),
	}

	return metrics
}

// NewCachedMetrics creates a new CachedMetrics with pre-cached common combinations
func NewCachedMetrics() *CachedMetrics {
	baseMetrics := NewPrometheusMetrics()

	// Pre-cache common HTTP request combinations
	deliveryRequests200, _ := baseMetrics.HTTPRequestsTotal.GetMetricWithLabelValues("GET", "/v1/delivery", "200")
	deliveryRequests400, _ := baseMetrics.HTTPRequestsTotal.GetMetricWithLabelValues("GET", "/v1/delivery", "400")
	deliveryRequests500, _ := baseMetrics.HTTPRequestsTotal.GetMetricWithLabelValues("GET", "/v1/delivery", "500")
	deliveryDuration, _ := baseMetrics.HTTPRequestDuration.GetMetricWithLabelValues("GET", "/v1/delivery")
	deliveryInFlight, _ := baseMetrics.HTTPRequestsInFlight.GetMetricWithLabelValues("GET", "/v1/delivery")

	// Pre-cache health endpoint combinations
	healthRequests200, _ := baseMetrics.HTTPRequestsTotal.GetMetricWithLabelValues("GET", "/health", "200")
	healthRequests500, _ := baseMetrics.HTTPRequestsTotal.GetMetricWithLabelValues("GET", "/health", "500")
	healthDuration, _ := baseMetrics.HTTPRequestDuration.GetMetricWithLabelValues("GET", "/health")
	healthInFlight, _ := baseMetrics.HTTPRequestsInFlight.GetMetricWithLabelValues("GET", "/health")

	// Pre-cache common database operations
	dbCampaignsSelect, _ := baseMetrics.DatabaseQueries.GetMetricWithLabelValues("select", "campaigns")
	dbTargetingRulesSelect, _ := baseMetrics.DatabaseQueries.GetMetricWithLabelValues("select", "targeting_rules")
	dbQueryError, _ := baseMetrics.DatabaseErrors.GetMetricWithLabelValues("select", "query_error")

	// Pre-cache health check statuses
	healthCheckDB, _ := baseMetrics.HealthCheckStatus.GetMetricWithLabelValues("database")
	healthCheckCache, _ := baseMetrics.HealthCheckStatus.GetMetricWithLabelValues("cache")

	return &CachedMetrics{
		Metrics: baseMetrics,

		// HTTP request caches
		deliveryRequests200: deliveryRequests200,
		deliveryRequests400: deliveryRequests400,
		deliveryRequests500: deliveryRequests500,
		deliveryDuration:    deliveryDuration,
		deliveryInFlight:    deliveryInFlight,

		healthRequests200: healthRequests200,
		healthRequests500: healthRequests500,
		healthDuration:    healthDuration,
		healthInFlight:    healthInFlight,

		// Database caches
		dbCampaignsSelect:      dbCampaignsSelect,
		dbTargetingRulesSelect: dbTargetingRulesSelect,
		dbQueryError:           dbQueryError,

		// Health check caches
		healthCheckDB:    healthCheckDB,
		healthCheckCache: healthCheckCache,
	}
}

// RecordHTTPRequest records an HTTP request with its duration and status
// Uses fast path for common combinations, falls back to original method for others
func (m *CachedMetrics) RecordHTTPRequest(method, endpoint, statusCode string, duration float64) {
	if method == "GET" && endpoint == "/v1/delivery" {
		m.deliveryDuration.Observe(duration)
		switch statusCode {
		case "200":
			m.deliveryRequests200.Inc()
			return
		case "400":
			m.deliveryRequests400.Inc()
			return
		case "500":
			m.deliveryRequests500.Inc()
			return
		}
	}

	if method == "GET" && endpoint == "/health" {
		m.healthDuration.Observe(duration)
		switch statusCode {
		case "200":
			m.healthRequests200.Inc()
			return
		case "500":
			m.healthRequests500.Inc()
			return
		}
	}

	// Fallback to original method for uncommon combinations
	m.Metrics.RecordHTTPRequest(method, endpoint, statusCode, duration)
}

// IncRequestsInFlight increments the in-flight requests counter
func (m *CachedMetrics) IncRequestsInFlight(method, endpoint string) {
	if method == "GET" && endpoint == "/v1/delivery" {
		m.deliveryInFlight.Inc()
		return
	}

	// Fast path for health endpoint
	if method == "GET" && endpoint == "/health" {
		m.healthInFlight.Inc()
		return
	}

	// Fallback to original method
	m.Metrics.IncRequestsInFlight(method, endpoint)
}

// DecRequestsInFlight decrements the in-flight requests counter
func (m *CachedMetrics) DecRequestsInFlight(method, endpoint string) {
	// Fast path for delivery endpoint
	if method == "GET" && endpoint == "/v1/delivery" {
		m.deliveryInFlight.Dec()
		return
	}

	// Fast path for health endpoint
	if method == "GET" && endpoint == "/health" {
		m.healthInFlight.Dec()
		return
	}

	// Fallback to original method
	m.Metrics.DecRequestsInFlight(method, endpoint)
}

// RecordDatabaseQuery records a database query
func (m *CachedMetrics) RecordDatabaseQuery(operation, table string) {
	if operation == "select" {
		switch table {
		case "campaigns":
			m.dbCampaignsSelect.Inc()
			return
		case "targeting_rules":
			m.dbTargetingRulesSelect.Inc()
			return
		}
	}

	// Fallback to original method
	m.Metrics.RecordDatabaseQuery(operation, table)
}

// RecordDatabaseError records a database error
func (m *CachedMetrics) RecordDatabaseError(operation, errorType string) {
	if operation == "select" && errorType == "query_error" {
		m.dbQueryError.Inc()
		return
	}

	// Fallback to original method
	m.Metrics.RecordDatabaseError(operation, errorType)
}

// SetHealthCheckStatus sets the health check status
func (m *CachedMetrics) SetHealthCheckStatus(checkType string, healthy bool) {
	status := 0.0
	if healthy {
		status = 1.0
	}

	switch checkType {
	case "database":
		m.healthCheckDB.Set(status)
		return
	case "cache":
		m.healthCheckCache.Set(status)
		return
	}

	// Fallback to original method
	m.Metrics.SetHealthCheckStatus(checkType, healthy)
}

// RecordCampaignDelivery records a campaign delivery
// This method doesn't need caching as it has many unique combinations
func (m *CachedMetrics) RecordCampaignDelivery(app, country, os string, count int) {
	m.Metrics.RecordCampaignDelivery(app, country, os, count)
}

// Original methods kept for backward compatibility
func (m *Metrics) RecordHTTPRequest(method, endpoint, statusCode string, duration float64) {
	m.HTTPRequestsTotal.WithLabelValues(method, endpoint, statusCode).Inc()
	m.HTTPRequestDuration.WithLabelValues(method, endpoint).Observe(duration)
}

func (m *Metrics) RecordCampaignDelivery(app, country, os string, count int) {
	m.CampaignsDelivered.WithLabelValues(app, country, os).Add(float64(count))
}

func (m *Metrics) RecordDatabaseQuery(operation, table string) {
	m.DatabaseQueries.WithLabelValues(operation, table).Inc()
}

func (m *Metrics) RecordDatabaseError(operation, errorType string) {
	m.DatabaseErrors.WithLabelValues(operation, errorType).Inc()
}

func (m *Metrics) SetHealthCheckStatus(checkType string, healthy bool) {
	status := 0.0
	if healthy {
		status = 1.0
	}
	m.HealthCheckStatus.WithLabelValues(checkType).Set(status)
}

func (m *Metrics) IncRequestsInFlight(method, endpoint string) {
	m.HTTPRequestsInFlight.WithLabelValues(method, endpoint).Inc()
}

func (m *Metrics) DecRequestsInFlight(method, endpoint string) {
	m.HTTPRequestsInFlight.WithLabelValues(method, endpoint).Dec()
}
