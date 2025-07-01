package transport

import (
	"context"
	"encoding/json"
	"net/http"

	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/go-kit/log"
	"github.com/gorilla/mux"
	"github.com/prajwalbharadwajbm/adbeacon/internal/cache"
	"github.com/prajwalbharadwajbm/adbeacon/internal/database"
	"github.com/prajwalbharadwajbm/adbeacon/internal/endpoint"
	"github.com/prajwalbharadwajbm/adbeacon/internal/models"
)

// NewHTTPHandler creates HTTP handlers for delivery service
func NewHTTPHandler(endpoints endpoint.DeliveryEndpoints, logger log.Logger) http.Handler {
	return NewHTTPHandlerWithDB(endpoints, logger, nil)
}

// NewHTTPHandlerWithDB creates HTTP handlers for delivery service with database health check
func NewHTTPHandlerWithDB(endpoints endpoint.DeliveryEndpoints, logger log.Logger, db *database.DB) http.Handler {
	return NewHTTPHandlerWithCache(endpoints, logger, db, nil)
}

// NewHTTPHandlerWithCache creates HTTP handlers with both database and cache health checks
func NewHTTPHandlerWithCache(endpoints endpoint.DeliveryEndpoints, logger log.Logger, db *database.DB, cache cache.Cache) http.Handler {
	options := []httptransport.ServerOption{
		httptransport.ServerErrorEncoder(encodeError),
	}

	getCampaignsHandler := httptransport.NewServer(
		endpoints.GetCampaignsEndpoint,
		decodeGetCampaignsRequest,
		encodeGetCampaignsResponse,
		options...,
	)

	r := mux.NewRouter()

	// Main delivery endpoint
	r.Handle("/v1/delivery", getCampaignsHandler).Methods("GET")

	// Health check endpoint with database and cache checks
	r.HandleFunc("/health", createHealthHandler(db, cache)).Methods("GET")

	return r
}

// decodeGetCampaignsRequest decodes HTTP request to GetCampaignsRequest
func decodeGetCampaignsRequest(_ context.Context, r *http.Request) (interface{}, error) {
	query := r.URL.Query()

	req := endpoint.GetCampaignsRequest{
		DeliveryRequest: models.DeliveryRequest{
			App:     query.Get("app"),
			Country: query.Get("country"),
			OS:      query.Get("os"),
		},
	}

	return req, nil
}

// encodeGetCampaignsResponse encodes GetCampaignsResponse to HTTP response
func encodeGetCampaignsResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	resp := response.(endpoint.GetCampaignsResponse)

	// Handle validation errors
	if resp.Err != nil {
		encodeError(ctx, resp.Err, w)
		return nil
	}

	// Handle empty results
	if len(resp.Campaigns) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return nil
	}

	// Return successful response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(resp.Campaigns)
}

// encodeError encodes error to HTTP response
func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")

	// Check for validation errors - these should return 400 Bad Request
	errorMsg := err.Error()
	if errorMsg == "country is required" ||
		errorMsg == "country must be a 2-letter code" ||
		errorMsg == "os is required" ||
		errorMsg == "app is required" ||
		errorMsg == "missing app param" ||
		errorMsg == "missing country param" ||
		errorMsg == "missing os param" {
		w.WriteHeader(http.StatusBadRequest)
	} else {
		// All other errors are internal server errors
		w.WriteHeader(http.StatusInternalServerError)
	}

	errorResponse := models.NewErrorResponse(err.Error())
	json.NewEncoder(w).Encode(errorResponse)
}

// createHealthHandler creates a health handler with optional database and cache checks
func createHealthHandler(db *database.DB, cache cache.Cache) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		response := map[string]any{
			"status":  "healthy",
			"service": "adbeacon",
			"version": "1.0.0",
		}

		overallHealthy := true

		// Check database health if available
		if db != nil {
			if err := db.HealthCheck(); err != nil {
				response["status"] = "unhealthy"
				response["database"] = map[string]any{
					"status": "unhealthy",
					"error":  err.Error(),
				}
				overallHealthy = false
			} else {
				// Add connection stats
				stats := db.GetConnectionStats()
				response["database"] = map[string]any{
					"status": "healthy",
					"stats": map[string]any{
						"open_connections":     stats.OpenConnections,
						"in_use":               stats.InUse,
						"idle":                 stats.Idle,
						"wait_count":           stats.WaitCount,
						"wait_duration":        stats.WaitDuration.String(),
						"max_idle_closed":      stats.MaxIdleClosed,
						"max_idle_time_closed": stats.MaxIdleTimeClosed,
						"max_lifetime_closed":  stats.MaxLifetimeClosed,
					},
				}
			}
		}

		// Check cache health if available
		if cache != nil {
			cacheHealth := cache.HealthCheck(ctx)
			response["cache"] = cacheHealth

			// Update overall status based on cache health
			if cacheHealth.Overall == "unhealthy" {
				response["status"] = "unhealthy"
				overallHealthy = false
			} else if cacheHealth.Overall == "degraded" && response["status"] != "unhealthy" {
				response["status"] = "degraded"
			}
		}

		// Set appropriate HTTP status code
		statusCode := http.StatusOK
		if !overallHealthy {
			statusCode = http.StatusServiceUnavailable
		} else if response["status"] == "degraded" {
			statusCode = http.StatusOK // 200 OK but with degraded status
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		json.NewEncoder(w).Encode(response)
	}
}
