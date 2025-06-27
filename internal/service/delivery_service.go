package service

import (
	"context"
	"errors"

	"github.com/prajwalbharadwajbm/adbeacon/internal/models"
)

// DeliveryService defines the interface for campaign delivery service
type DeliveryService interface {
	GetCampaigns(ctx context.Context, req models.DeliveryRequest) ([]models.CampaignResponse, error)
}

// deliveryService implements DeliveryService interface
type deliveryService struct {
	campaignRepo CampaignRepository
}

// CampaignRepository interface for data access
type CampaignRepository interface {
	GetActiveCampaignsWithRules(ctx context.Context) ([]models.CampaignWithRules, error)
}

// NewDeliveryService creates a new delivery service
func NewDeliveryService(campaignRepo CampaignRepository) DeliveryService {
	return &deliveryService{
		campaignRepo: campaignRepo,
	}
}

// GetCampaigns finds all campaigns that match the delivery request
func (s *deliveryService) GetCampaigns(ctx context.Context, req models.DeliveryRequest) ([]models.CampaignResponse, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Normalize request for consistent matching
	req.Normalize()

	// Get all active campaigns with their targeting rules
	campaignsWithRules, err := s.campaignRepo.GetActiveCampaignsWithRules(ctx)
	if err != nil {
		return nil, errors.New("failed to retrieve campaigns")
	}

	// Find matching campaigns
	var matchingCampaigns []models.CampaignResponse
	for _, campaign := range campaignsWithRules {
		if campaign.MatchesRequest(req) {
			matchingCampaigns = append(matchingCampaigns, campaign.ToResponse())
		}
	}

	return matchingCampaigns, nil
}
