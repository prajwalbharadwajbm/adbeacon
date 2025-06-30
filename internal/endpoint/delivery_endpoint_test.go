package endpoint

import (
	"context"
	"errors"
	"testing"

	"github.com/prajwalbharadwajbm/adbeacon/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockDeliveryService is a mock implementation of service.DeliveryService
type MockDeliveryService struct {
	mock.Mock
}

func (m *MockDeliveryService) GetCampaigns(ctx context.Context, req models.DeliveryRequest) ([]models.CampaignResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).([]models.CampaignResponse), args.Error(1)
}

func TestMakeDeliveryEndpoints(t *testing.T) {
	mockService := &MockDeliveryService{}
	endpoints := MakeDeliveryEndpoints(mockService)

	assert.NotNil(t, endpoints)
	assert.NotNil(t, endpoints.GetCampaignsEndpoint)
}

func TestGetCampaignsEndpoint_Success(t *testing.T) {
	mockService := &MockDeliveryService{}
	endpoints := MakeDeliveryEndpoints(mockService)

	// Setup mock response
	expectedCampaigns := []models.CampaignResponse{
		{CID: "spotify", Img: "https://example.com/spotify.jpg", CTA: "Download"},
		{CID: "duolingo", Img: "https://example.com/duolingo.jpg", CTA: "Install"},
	}

	mockService.On("GetCampaigns", mock.Anything, mock.MatchedBy(func(req models.DeliveryRequest) bool {
		return req.App == "com.test.app" && req.Country == "US" && req.OS == "Android"
	})).Return(expectedCampaigns, nil)

	// Create request
	request := GetCampaignsRequest{
		DeliveryRequest: models.DeliveryRequest{
			App:     "com.test.app",
			Country: "US",
			OS:      "Android",
		},
	}

	// Call endpoint
	response, err := endpoints.GetCampaignsEndpoint(context.Background(), request)

	assert.NoError(t, err)
	assert.IsType(t, GetCampaignsResponse{}, response)

	getCampaignsResponse := response.(GetCampaignsResponse)
	assert.Equal(t, expectedCampaigns, getCampaignsResponse.Campaigns)
	assert.Nil(t, getCampaignsResponse.Err)

	mockService.AssertExpectations(t)
}

func TestGetCampaignsEndpoint_NoResults(t *testing.T) {
	mockService := &MockDeliveryService{}
	endpoints := MakeDeliveryEndpoints(mockService)

	// Setup mock to return empty results
	mockService.On("GetCampaigns", mock.Anything, mock.Anything).Return([]models.CampaignResponse{}, nil)

	request := GetCampaignsRequest{
		DeliveryRequest: models.DeliveryRequest{
			App:     "com.test.app",
			Country: "CA",
			OS:      "iOS",
		},
	}

	response, err := endpoints.GetCampaignsEndpoint(context.Background(), request)

	assert.NoError(t, err)
	assert.IsType(t, GetCampaignsResponse{}, response)

	getCampaignsResponse := response.(GetCampaignsResponse)
	assert.Empty(t, getCampaignsResponse.Campaigns)
	assert.Nil(t, getCampaignsResponse.Err)

	mockService.AssertExpectations(t)
}

func TestGetCampaignsEndpoint_ServiceError(t *testing.T) {
	mockService := &MockDeliveryService{}
	endpoints := MakeDeliveryEndpoints(mockService)

	// Setup mock to return an error
	serviceError := errors.New("service error")
	mockService.On("GetCampaigns", mock.Anything, mock.Anything).Return([]models.CampaignResponse{}, serviceError)

	request := GetCampaignsRequest{
		DeliveryRequest: models.DeliveryRequest{
			App:     "com.test.app",
			Country: "US",
			OS:      "Android",
		},
	}

	response, err := endpoints.GetCampaignsEndpoint(context.Background(), request)

	assert.NoError(t, err) // Endpoint itself doesn't return error, error is in response
	assert.IsType(t, GetCampaignsResponse{}, response)

	getCampaignsResponse := response.(GetCampaignsResponse)
	assert.Empty(t, getCampaignsResponse.Campaigns)
	assert.Equal(t, serviceError, getCampaignsResponse.Err)

	mockService.AssertExpectations(t)
}

func TestGetCampaignsEndpoint_ValidationError(t *testing.T) {
	mockService := &MockDeliveryService{}
	endpoints := MakeDeliveryEndpoints(mockService)

	// Setup mock to return validation error
	validationError := errors.New("missing app param")
	mockService.On("GetCampaigns", mock.Anything, mock.Anything).Return([]models.CampaignResponse{}, validationError)

	request := GetCampaignsRequest{
		DeliveryRequest: models.DeliveryRequest{
			App:     "",
			Country: "US",
			OS:      "Android",
		},
	}

	response, err := endpoints.GetCampaignsEndpoint(context.Background(), request)

	assert.NoError(t, err)
	assert.IsType(t, GetCampaignsResponse{}, response)

	getCampaignsResponse := response.(GetCampaignsResponse)
	assert.Empty(t, getCampaignsResponse.Campaigns)
	assert.Equal(t, validationError, getCampaignsResponse.Err)

	mockService.AssertExpectations(t)
}

func TestGetCampaignsResponse_Failed(t *testing.T) {
	// Test the Failed() method implementation
	response := &GetCampaignsResponse{
		Campaigns: []models.CampaignResponse{},
		Err:       errors.New("test error"),
	}

	assert.NotNil(t, response.Failed())
	assert.Equal(t, "test error", response.Failed().Error())

	// Test with no error
	responseNoError := &GetCampaignsResponse{
		Campaigns: []models.CampaignResponse{},
		Err:       nil,
	}

	assert.Nil(t, responseNoError.Failed())
}

func TestGetCampaignsEndpoint_ContextHandling(t *testing.T) {
	mockService := &MockDeliveryService{}
	endpoints := MakeDeliveryEndpoints(mockService)

	// Test that context is properly passed to service
	mockService.On("GetCampaigns", mock.MatchedBy(func(ctx context.Context) bool {
		return ctx.Value("test-key") == "test-value"
	}), mock.Anything).Return([]models.CampaignResponse{}, nil)

	// Create context with value
	ctx := context.WithValue(context.Background(), "test-key", "test-value")

	request := GetCampaignsRequest{
		DeliveryRequest: models.DeliveryRequest{
			App:     "com.test.app",
			Country: "US",
			OS:      "Android",
		},
	}

	_, err := endpoints.GetCampaignsEndpoint(ctx, request)

	assert.NoError(t, err)
	mockService.AssertExpectations(t)
}

func TestGetCampaignsEndpoint_RequestTransformation(t *testing.T) {
	mockService := &MockDeliveryService{}
	endpoints := MakeDeliveryEndpoints(mockService)

	// Test that the endpoint request is properly transformed to service request
	mockService.On("GetCampaigns", mock.Anything, mock.MatchedBy(func(req models.DeliveryRequest) bool {
		return req.App == "com.test.app" &&
			req.Country == "US" &&
			req.OS == "Android"
	})).Return([]models.CampaignResponse{}, nil)

	request := GetCampaignsRequest{
		DeliveryRequest: models.DeliveryRequest{
			App:     "com.test.app",
			Country: "US",
			OS:      "Android",
		},
	}

	_, err := endpoints.GetCampaignsEndpoint(context.Background(), request)

	assert.NoError(t, err)
	mockService.AssertExpectations(t)
}
