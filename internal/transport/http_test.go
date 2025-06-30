package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-kit/log"
	"github.com/gorilla/mux"
	"github.com/prajwalbharadwajbm/adbeacon/internal/endpoint"
	"github.com/prajwalbharadwajbm/adbeacon/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockEndpoints mocks the endpoint.DeliveryEndpoints
type MockEndpoints struct {
	mock.Mock
}

func (m *MockEndpoints) GetCampaignsEndpoint(ctx context.Context, request interface{}) (interface{}, error) {
	args := m.Called(ctx, request)
	return args.Get(0), args.Error(1)
}

func TestNewHTTPHandler(t *testing.T) {
	logger := log.NewNopLogger()
	endpoints := endpoint.DeliveryEndpoints{}

	handler := NewHTTPHandler(endpoints, logger)

	assert.NotNil(t, handler)
	assert.IsType(t, &mux.Router{}, handler)
}

func TestHealthEndpoint(t *testing.T) {
	logger := log.NewNopLogger()
	endpoints := endpoint.DeliveryEndpoints{}
	handler := NewHTTPHandler(endpoints, logger)

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "adbeacon", response["service"])
	assert.Equal(t, "healthy", response["status"])
	assert.Equal(t, "1.0.0", response["version"])
}

func TestDecodeGetCampaignsRequest_Success(t *testing.T) {
	values := url.Values{}
	values.Set("app", "com.test.app")
	values.Set("country", "US")
	values.Set("os", "Android")

	req := httptest.NewRequest("GET", "/v1/delivery?"+values.Encode(), nil)

	result, err := decodeGetCampaignsRequest(context.Background(), req)

	assert.NoError(t, err)
	assert.IsType(t, endpoint.GetCampaignsRequest{}, result)

	getCampaignsReq := result.(endpoint.GetCampaignsRequest)
	assert.Equal(t, "com.test.app", getCampaignsReq.DeliveryRequest.App)
	assert.Equal(t, "US", getCampaignsReq.DeliveryRequest.Country)
	assert.Equal(t, "Android", getCampaignsReq.DeliveryRequest.OS)
}

func TestDecodeGetCampaignsRequest_MissingParams(t *testing.T) {
	tests := []struct {
		name        string
		queryParams url.Values
		wantErr     string
	}{
		{
			name:        "missing app",
			queryParams: url.Values{"country": {"US"}, "os": {"Android"}},
			wantErr:     "missing app param",
		},
		{
			name:        "missing country",
			queryParams: url.Values{"app": {"com.test.app"}, "os": {"Android"}},
			wantErr:     "missing country param",
		},
		{
			name:        "missing os",
			queryParams: url.Values{"app": {"com.test.app"}, "country": {"US"}},
			wantErr:     "missing os param",
		},
		{
			name:        "all missing",
			queryParams: url.Values{},
			wantErr:     "missing app param",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/v1/delivery?"+tt.queryParams.Encode(), nil)

			// The decode function doesn't validate - validation happens in service layer
			result, err := decodeGetCampaignsRequest(context.Background(), req)

			// Decode should succeed but create request with missing fields
			assert.NoError(t, err)
			getCampaignsReq := result.(endpoint.GetCampaignsRequest)

			// Verify the request has empty fields (which will fail validation later)
			switch tt.wantErr {
			case "missing app param":
				assert.Empty(t, getCampaignsReq.DeliveryRequest.App)
			case "missing country param":
				assert.Empty(t, getCampaignsReq.DeliveryRequest.Country)
			case "missing os param":
				assert.Empty(t, getCampaignsReq.DeliveryRequest.OS)
			}
		})
	}
}

func TestEncodeGetCampaignsResponse_Success(t *testing.T) {
	campaigns := []models.CampaignResponse{
		{CID: "spotify", Img: "https://example.com/spotify.jpg", CTA: "Download"},
		{CID: "duolingo", Img: "https://example.com/duolingo.jpg", CTA: "Install"},
	}

	response := endpoint.GetCampaignsResponse{
		Campaigns: campaigns,
		Err:       nil,
	}

	w := httptest.NewRecorder()
	err := encodeGetCampaignsResponse(context.Background(), w, response)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var decodedCampaigns []models.CampaignResponse
	err = json.Unmarshal(w.Body.Bytes(), &decodedCampaigns)
	assert.NoError(t, err)
	assert.Equal(t, campaigns, decodedCampaigns)
}

func TestEncodeGetCampaignsResponse_EmptyResults(t *testing.T) {
	response := endpoint.GetCampaignsResponse{
		Campaigns: []models.CampaignResponse{},
		Err:       nil,
	}

	w := httptest.NewRecorder()
	err := encodeGetCampaignsResponse(context.Background(), w, response)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Empty(t, w.Body.String())
}

func TestEncodeGetCampaignsResponse_ValidationError(t *testing.T) {
	response := endpoint.GetCampaignsResponse{
		Campaigns: []models.CampaignResponse{},
		Err:       errors.New("missing app param"),
	}

	w := httptest.NewRecorder()
	err := encodeGetCampaignsResponse(context.Background(), w, response)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var errorResponse models.ErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)
	assert.Equal(t, "missing app param", errorResponse.Error)
}

func TestEncodeGetCampaignsResponse_InternalError(t *testing.T) {
	response := endpoint.GetCampaignsResponse{
		Campaigns: []models.CampaignResponse{},
		Err:       errors.New("database connection failed"),
	}

	w := httptest.NewRecorder()
	err := encodeGetCampaignsResponse(context.Background(), w, response)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var errorResponse models.ErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)
	assert.Equal(t, "database connection failed", errorResponse.Error)
}

func TestDeliveryEndpoint_Integration(t *testing.T) {
	logger := log.NewNopLogger()

	// Create mock endpoints
	mockEndpoints := &MockEndpoints{}

	// Setup mock response
	expectedCampaigns := []models.CampaignResponse{
		{CID: "spotify", Img: "https://example.com/spotify.jpg", CTA: "Download"},
	}

	mockEndpoints.On("GetCampaignsEndpoint", mock.Anything, mock.MatchedBy(func(req endpoint.GetCampaignsRequest) bool {
		return req.DeliveryRequest.App == "com.test.app" && req.DeliveryRequest.Country == "US" && req.DeliveryRequest.OS == "Android"
	})).Return(endpoint.GetCampaignsResponse{
		Campaigns: expectedCampaigns,
		Err:       nil,
	}, nil)

	// Create handler with mock endpoints
	endpoints := endpoint.DeliveryEndpoints{
		GetCampaignsEndpoint: mockEndpoints.GetCampaignsEndpoint,
	}
	handler := NewHTTPHandler(endpoints, logger)

	// Make request
	req := httptest.NewRequest("GET", "/v1/delivery?app=com.test.app&country=US&os=Android", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var campaigns []models.CampaignResponse
	err := json.Unmarshal(w.Body.Bytes(), &campaigns)
	assert.NoError(t, err)
	assert.Equal(t, expectedCampaigns, campaigns)

	mockEndpoints.AssertExpectations(t)
}

func TestDeliveryEndpoint_ValidationError_Integration(t *testing.T) {
	logger := log.NewNopLogger()

	mockEndpoints := &MockEndpoints{}
	mockEndpoints.On("GetCampaignsEndpoint", mock.Anything, mock.Anything).Return(endpoint.GetCampaignsResponse{
		Campaigns: []models.CampaignResponse{},
		Err:       errors.New("missing country param"),
	}, nil)

	endpoints := endpoint.DeliveryEndpoints{
		GetCampaignsEndpoint: mockEndpoints.GetCampaignsEndpoint,
	}
	handler := NewHTTPHandler(endpoints, logger)

	req := httptest.NewRequest("GET", "/v1/delivery?app=com.test.app&os=Android", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var errorResponse models.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)
	assert.Contains(t, errorResponse.Error, "missing country param")

	mockEndpoints.AssertExpectations(t)
}

func TestDeliveryEndpoint_NoResults_Integration(t *testing.T) {
	logger := log.NewNopLogger()

	mockEndpoints := &MockEndpoints{}
	mockEndpoints.On("GetCampaignsEndpoint", mock.Anything, mock.Anything).Return(endpoint.GetCampaignsResponse{
		Campaigns: []models.CampaignResponse{},
		Err:       nil,
	}, nil)

	endpoints := endpoint.DeliveryEndpoints{
		GetCampaignsEndpoint: mockEndpoints.GetCampaignsEndpoint,
	}
	handler := NewHTTPHandler(endpoints, logger)

	req := httptest.NewRequest("GET", "/v1/delivery?app=com.test.app&country=CA&os=iOS", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Empty(t, w.Body.String())

	mockEndpoints.AssertExpectations(t)
}

func TestHTTPHandler_MethodNotAllowed(t *testing.T) {
	logger := log.NewNopLogger()
	endpoints := endpoint.DeliveryEndpoints{}
	handler := NewHTTPHandler(endpoints, logger)

	// Test POST to delivery endpoint (should be GET only)
	req := httptest.NewRequest("POST", "/v1/delivery", bytes.NewBuffer([]byte("{}")))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestHTTPHandler_NotFound(t *testing.T) {
	logger := log.NewNopLogger()
	endpoints := endpoint.DeliveryEndpoints{}
	handler := NewHTTPHandler(endpoints, logger)

	req := httptest.NewRequest("GET", "/nonexistent", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}
