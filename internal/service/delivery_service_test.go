package service

import (
	"context"
	"errors"
	"testing"

	"github.com/prajwalbharadwajbm/adbeacon/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockCampaignRepository is a mock implementation of CampaignRepository
type MockCampaignRepository struct {
	mock.Mock
}

func (m *MockCampaignRepository) GetActiveCampaignsWithRules(ctx context.Context) ([]models.CampaignWithRules, error) {
	args := m.Called(ctx)
	return args.Get(0).([]models.CampaignWithRules), args.Error(1)
}

func TestNewDeliveryService(t *testing.T) {
	mockRepo := &MockCampaignRepository{}
	service := NewDeliveryService(mockRepo)

	assert.NotNil(t, service)
	assert.IsType(t, &deliveryService{}, service)
}

func TestDeliveryService_GetCampaigns_InvalidRequest(t *testing.T) {
	mockRepo := &MockCampaignRepository{}
	service := NewDeliveryService(mockRepo)

	tests := []struct {
		name    string
		request models.DeliveryRequest
		wantErr string
	}{
		{
			name: "missing app",
			request: models.DeliveryRequest{
				Country: "US",
				OS:      "Android",
			},
			wantErr: "missing app param",
		},
		{
			name: "missing country",
			request: models.DeliveryRequest{
				App: "com.test.app",
				OS:  "Android",
			},
			wantErr: "missing country param",
		},
		{
			name: "missing os",
			request: models.DeliveryRequest{
				App:     "com.test.app",
				Country: "US",
			},
			wantErr: "missing os param",
		},
		{
			name:    "all missing",
			request: models.DeliveryRequest{},
			wantErr: "missing app param",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.GetCampaigns(context.Background(), tt.request)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestDeliveryService_GetCampaigns_RepositoryError(t *testing.T) {
	mockRepo := &MockCampaignRepository{}
	service := NewDeliveryService(mockRepo)

	// Setup mock to return an error
	mockRepo.On("GetActiveCampaignsWithRules", mock.Anything).Return([]models.CampaignWithRules{}, errors.New("database error"))

	request := models.DeliveryRequest{
		App:     "com.test.app",
		Country: "US",
		OS:      "Android",
	}

	_, err := service.GetCampaigns(context.Background(), request)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to retrieve campaigns")

	mockRepo.AssertExpectations(t)
}

func TestDeliveryService_GetCampaigns_NoMatchingCampaigns(t *testing.T) {
	mockRepo := &MockCampaignRepository{}
	service := NewDeliveryService(mockRepo)

	// Setup mock with campaigns that don't match the request
	campaigns := []models.CampaignWithRules{
		createTestCampaign("test1", models.StatusActive, []models.TargetingRule{
			{
				Dimension: models.DimensionCountry,
				RuleType:  models.RuleTypeInclude,
				Values:    []string{"CA"}, // Request is for US
			},
		}),
	}

	mockRepo.On("GetActiveCampaignsWithRules", mock.Anything).Return(campaigns, nil)

	request := models.DeliveryRequest{
		App:     "com.test.app",
		Country: "US",
		OS:      "Android",
	}

	result, err := service.GetCampaigns(context.Background(), request)
	assert.NoError(t, err)
	assert.Empty(t, result)

	mockRepo.AssertExpectations(t)
}

func TestDeliveryService_GetCampaigns_MatchingCampaigns(t *testing.T) {
	mockRepo := &MockCampaignRepository{}
	service := NewDeliveryService(mockRepo)

	// Setup mock with matching campaigns
	campaigns := []models.CampaignWithRules{
		createTestCampaign("spotify", models.StatusActive, []models.TargetingRule{
			{
				Dimension: models.DimensionCountry,
				RuleType:  models.RuleTypeInclude,
				Values:    []string{"US", "CA"},
			},
		}),
		createTestCampaign("duolingo", models.StatusActive, []models.TargetingRule{
			{
				Dimension: models.DimensionOS,
				RuleType:  models.RuleTypeInclude,
				Values:    []string{"Android", "iOS"},
			},
			{
				Dimension: models.DimensionCountry,
				RuleType:  models.RuleTypeExclude,
				Values:    []string{"US"}, // This should exclude it
			},
		}),
		createTestCampaign("subwaysurfer", models.StatusActive, []models.TargetingRule{
			{
				Dimension: models.DimensionOS,
				RuleType:  models.RuleTypeInclude,
				Values:    []string{"Android"},
			},
			{
				Dimension: models.DimensionApp,
				RuleType:  models.RuleTypeInclude,
				Values:    []string{"com.test.app"},
			},
		}),
	}

	mockRepo.On("GetActiveCampaignsWithRules", mock.Anything).Return(campaigns, nil)

	request := models.DeliveryRequest{
		App:     "com.test.app",
		Country: "US",
		OS:      "Android",
	}

	result, err := service.GetCampaigns(context.Background(), request)
	assert.NoError(t, err)
	assert.Len(t, result, 2) // spotify and subwaysurfer should match, duolingo excluded

	// Check that we got the right campaigns
	campaignIDs := make([]string, len(result))
	for i, campaign := range result {
		campaignIDs[i] = campaign.CID
	}
	assert.Contains(t, campaignIDs, "spotify")
	assert.Contains(t, campaignIDs, "subwaysurfer")
	assert.NotContains(t, campaignIDs, "duolingo")

	mockRepo.AssertExpectations(t)
}

func TestDeliveryService_GetCampaigns_InactiveCampaignsFiltered(t *testing.T) {
	mockRepo := &MockCampaignRepository{}
	service := NewDeliveryService(mockRepo)

	// Setup mock with inactive campaigns (should be filtered out by repository)
	campaigns := []models.CampaignWithRules{
		createTestCampaign("active", models.StatusActive, []models.TargetingRule{
			{
				Dimension: models.DimensionCountry,
				RuleType:  models.RuleTypeInclude,
				Values:    []string{"US"},
			},
		}),
		// Note: Inactive campaigns should already be filtered out by repository
	}

	mockRepo.On("GetActiveCampaignsWithRules", mock.Anything).Return(campaigns, nil)

	request := models.DeliveryRequest{
		App:     "com.test.app",
		Country: "US",
		OS:      "Android",
	}

	result, err := service.GetCampaigns(context.Background(), request)
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "active", result[0].CID)

	mockRepo.AssertExpectations(t)
}

func TestDeliveryService_GetCampaigns_RequestNormalization(t *testing.T) {
	mockRepo := &MockCampaignRepository{}
	service := NewDeliveryService(mockRepo)

	campaigns := []models.CampaignWithRules{
		createTestCampaign("test", models.StatusActive, []models.TargetingRule{
			{
				Dimension: models.DimensionCountry,
				RuleType:  models.RuleTypeInclude,
				Values:    []string{"us"}, // lowercase in data
			},
		}),
	}

	mockRepo.On("GetActiveCampaignsWithRules", mock.Anything).Return(campaigns, nil)

	// Test with uppercase country - should be normalized to lowercase
	request := models.DeliveryRequest{
		App:     "  com.test.app  ", // with spaces
		Country: "US",               // uppercase
		OS:      "ANDROID",          // uppercase
	}

	result, err := service.GetCampaigns(context.Background(), request)
	assert.NoError(t, err)
	assert.Len(t, result, 1)

	mockRepo.AssertExpectations(t)
}

// Helper function to create test campaigns
func createTestCampaign(id string, status models.CampaignStatus, rules []models.TargetingRule) models.CampaignWithRules {
	return models.CampaignWithRules{
		Campaign: models.Campaign{
			ID:       id,
			Name:     id + " campaign",
			ImageURL: "https://example.com/" + id + ".jpg",
			CTA:      "Install",
			Status:   status,
		},
		Rules: rules,
	}
}
