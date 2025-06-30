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

// RecordHTTPRequest records an HTTP request with its duration and status
func (m *Metrics) RecordHTTPRequest(method, endpoint, statusCode string, duration float64) {
	m.HTTPRequestsTotal.WithLabelValues(method, endpoint, statusCode).Inc()
	m.HTTPRequestDuration.WithLabelValues(method, endpoint).Observe(duration)
}

// RecordCampaignDelivery records a campaign delivery
func (m *Metrics) RecordCampaignDelivery(app, country, os string, count int) {
	m.CampaignsDelivered.WithLabelValues(app, country, os).Add(float64(count))
}

// RecordDatabaseQuery records a database query
func (m *Metrics) RecordDatabaseQuery(operation, table string) {
	m.DatabaseQueries.WithLabelValues(operation, table).Inc()
}

// RecordDatabaseError records a database error
func (m *Metrics) RecordDatabaseError(operation, errorType string) {
	m.DatabaseErrors.WithLabelValues(operation, errorType).Inc()
}

// SetHealthCheckStatus sets the health check status
func (m *Metrics) SetHealthCheckStatus(checkType string, healthy bool) {
	status := 0.0
	if healthy {
		status = 1.0
	}
	m.HealthCheckStatus.WithLabelValues(checkType).Set(status)
}

// IncRequestsInFlight increments the in-flight requests counter
func (m *Metrics) IncRequestsInFlight(method, endpoint string) {
	m.HTTPRequestsInFlight.WithLabelValues(method, endpoint).Inc()
}

// DecRequestsInFlight decrements the in-flight requests counter
func (m *Metrics) DecRequestsInFlight(method, endpoint string) {
	m.HTTPRequestsInFlight.WithLabelValues(method, endpoint).Dec()
}
