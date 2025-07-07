package middleware

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/prajwalbharadwajbm/adbeacon/internal/metrics"
)

// MetricsMiddleware wraps HTTP handlers to collect Prometheus metrics
type MetricsMiddleware struct {
	metrics *metrics.CachedMetrics
}

// NewMetricsMiddleware creates a new metrics middleware
func NewMetricsMiddleware(metrics *metrics.CachedMetrics) *MetricsMiddleware {
	return &MetricsMiddleware{
		metrics: metrics,
	}
}

// Middleware returns the HTTP middleware function
func (m *MetricsMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Normalize endpoint path for metrics (remove query params and IDs)
		endpoint := normalizeEndpoint(r.URL.Path)
		method := r.Method

		// Increment in-flight requests
		m.metrics.IncRequestsInFlight(method, endpoint)
		defer m.metrics.DecRequestsInFlight(method, endpoint)

		// Wrap the response writer to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: 200}

		// Process the request
		next.ServeHTTP(wrapped, r)

		// Record metrics
		duration := time.Since(start).Seconds()
		statusCode := strconv.Itoa(wrapped.statusCode)

		m.metrics.RecordHTTPRequest(method, endpoint, statusCode, duration)
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

func (rw *responseWriter) WriteHeader(code int) {
	if !rw.written {
		rw.statusCode = code
		rw.written = true
	}
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.written {
		rw.WriteHeader(200)
	}
	return rw.ResponseWriter.Write(b)
}

// normalizeEndpoint normalizes URL paths for consistent metric labels
func normalizeEndpoint(path string) string {
	// Remove trailing slash
	path = strings.TrimSuffix(path, "/")

	// Handle common endpoints
	switch {
	case path == "/health":
		return "/health"
	case path == "/metrics":
		return "/metrics"
	case strings.HasPrefix(path, "/v1/delivery"):
		return "/v1/delivery"
	default:
		return path
	}
}
