package transport

import (
	"context"
	"encoding/json"
	"net/http"

	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/go-kit/log"
	"github.com/gorilla/mux"
	"github.com/prajwalbharadwajbm/adbeacon/internal/endpoint"
	"github.com/prajwalbharadwajbm/adbeacon/internal/models"
)

// NewHTTPHandler creates HTTP handlers for delivery service
func NewHTTPHandler(endpoints endpoint.DeliveryEndpoints, logger log.Logger) http.Handler {
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

	// Health check endpoint
	r.HandleFunc("/health", healthHandler).Methods("GET")

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

	// Check for validation errors
	if err.Error() == "missing app param" ||
		err.Error() == "missing country param" ||
		err.Error() == "missing os param" {
		w.WriteHeader(http.StatusBadRequest)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}

	errorResponse := models.NewErrorResponse(err.Error())
	json.NewEncoder(w).Encode(errorResponse)
}

// healthHandler handles health check requests
func healthHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]any{
		"status":  "healthy",
		"service": "adbeacon",
		"version": "1.0.0",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
